// WIT Room Scrape - Finds when rooms in the SETU Waterford campus are empty.
// Copyright (C) 2024 Piotr Placzek (piotrpdev)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/olekukonko/tablewriter"

	"net/http"
	"net/url"
)

type Timetable struct {
	Error     bool
	Empty     bool // Sanity check
	Date      string
	Room      string
	Week      string
	FreeTimes map[string][]string
}

const roomTimetableUrl = "http://studentssp.wit.ie/Timetables/RoomTT.aspx"

func setupLogger(filePath string, debugMode bool) (*os.File, error) {
	// ? https://stackoverflow.com/a/13513490/19020549
	f1, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening log file: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if debugMode {
		opts.Level = slog.LevelDebug
	}

	mw := io.MultiWriter(os.Stdout, f1)
	logger := slog.New(slog.NewTextHandler(mw, opts))
	slog.SetDefault(logger)

	return f1, nil
}

// https://gist.github.com/rustyeddy/77f17f4f0fb83cc87115eb72a23f18f7?permalink_comment_id=4069054#gistcomment-4069054
func getTimeStamp() string {
	ts := time.Now().UTC().Format(time.RFC3339)
	return strings.Replace(strings.Replace(ts, ":", "_", -1), "-", "_", -1)
}

// https://stackoverflow.com/a/18203895/19020549
func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// https://stackoverflow.com/a/20397167/19020549
func Upload(url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func sendAsciiFileToWebhook(asciiTable string, webhookUrl string) {
	// TODO: Convert to image because file is too large for Discord
	// `https://stackoverflow.com/a/38300583/19020549`

	var sb strings.Builder
	sb.WriteString("```")
	sb.WriteString(asciiTable)
	sb.WriteString("```")

	values := map[string]io.Reader{
		"username":   strings.NewReader("Peter's Room Checker Bot"),
		"avatar_url": strings.NewReader("https://i.imgur.com/oBPXx0D.png"),
		"content":    strings.NewReader(sb.String()),
	}

	err := Upload(webhookUrl, values)
	if err != nil {
		slog.Error(err.Error())
		slog.Error("Failed to send asciiTable to webhook")
		os.Exit(1)
	}
}

func main() {
	logPath := flag.String("logPath", "./room_scrape.log", "Path to log file")
	debugMode := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	slog.Info("Starting room-scrape", slog.String("logPath", *logPath), slog.Bool("debugMode", *debugMode))
	slog.Info("Setting up logger", slog.String("logPath", *logPath))

	f1, err := setupLogger(*logPath, *debugMode)
	if err != nil {
		slog.Error(err.Error())
		slog.Error("room-scrape failed to start, exitting")
		os.Exit(1)
	}

	defer f1.Close()

	slog.Info("room-scrape started successfully", slog.String("logPath", *logPath), slog.Bool("debugMode", *debugMode))

	// https://go-colly.org/docs/examples/coursera_courses/
	// roomRegexp, _ := regexp.Compile("^IT.{3}$")
	// roomRegexp, _ := regexp.Compile("^IT10[1|2]$")
	roomRegexp, _ := regexp.Compile("^IT1[0-9]{2}$")
	// roomRegexp, _ := regexp.Compile("^IT101$")

	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	times := []string{"09:15", "10:15", "11:15", "12:15", "13:15", "14:15", "15:15", "16:15", "17:15"}
	// TODO: Change to 0 in prod
	// weekOffset := -1
	weekOffset := 0

	timetableList := []Timetable{}

	initCollector := colly.NewCollector(
		colly.Async(false),
	)
	initCollector.Limit(&colly.LimitRule{Parallelism: 1})
	initCollector.SetRequestTimeout(time.Duration(30) * time.Second)

	roomCollector := initCollector.Clone()

	roomPostQueryParams := url.Values{}
	roomPostQueryParams.Add("__EVENTTARGET", "")
	roomPostQueryParams.Add("__EVENTARGUMENT", "")
	roomPostQueryParams.Add("__LASTFOCUS", "")
	roomPostQueryParams.Add("hProgram", "")
	roomPostQueryParams.Add("hStudentcount", "")
	roomPostQueryParams.Add("cboSchool", "%")
	roomPostQueryParams.Add("CboDept", "%")
	roomPostQueryParams.Add("CboStartTime", "1") // This is the default but might change in the future
	roomPostQueryParams.Add("CboEndTime", "9")   // This is the default but might change in the future
	roomPostQueryParams.Add("BtnRetrieve", "Generate Timetable")

	initCollector.OnHTML("body", func(bodyElement *colly.HTMLElement) {
		bodyElement.ForEach("input[type='hidden']", func(_ int, hiddenInputElement *colly.HTMLElement) {
			switch hiddenInputElement.Attr("name") {
			case "__VIEWSTATE":
				roomPostQueryParams.Add("__VIEWSTATE", hiddenInputElement.Attr("value"))
			case "__VIEWSTATEGENERATOR":
				roomPostQueryParams.Add("__VIEWSTATEGENERATOR", hiddenInputElement.Attr("value"))
			case "__EVENTVALIDATION":
				roomPostQueryParams.Add("__EVENTVALIDATION", hiddenInputElement.Attr("value"))
			}
		})

		// roomPostQueryParams.Add("CboWeeks", "34") // This works btw
		week := bodyElement.ChildAttr("select[name='CboWeeks'] > option[selected='selected']", "value")
		weekInt, err := strconv.Atoi(week)
		if err != nil {
			slog.Error(err.Error())
			slog.Error("Failed to convert week string to int", slog.String("week", week))
			roomPostQueryParams.Add("CboWeeks", week)
		} else {
			roomPostQueryParams.Add("CboWeeks", strconv.Itoa(weekInt+weekOffset))
		}

		bodyElement.ForEach("select[name='CboLocation'] > option", func(_ int, roomOptionElement *colly.HTMLElement) {
			roomName := roomOptionElement.Attr("value")

			if roomRegexp.MatchString(roomName) {
				slog.Debug("Room found", slog.String("roomName", roomName))

				roomPostQueryParams.Set("CboLocation", roomName)

				err := roomCollector.Request(
					"POST",
					roomTimetableUrl,
					strings.NewReader(roomPostQueryParams.Encode()),
					nil,
					http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
				)

				if err != nil {
					slog.Error(err.Error())
					slog.Error("Room POST request failed")
					return
				}
			}
		})
	})

	initCollector.OnRequest(func(r *colly.Request) {
		slog.Info("Getting required metadata", slog.String("URL", r.URL.String()))
	})

	initCollector.OnError(func(r *colly.Response, err error) {
		slog.Error(err.Error())
		slog.Error("Getting required metadata failed", slog.String("URL", r.Request.URL.String()), slog.Any("response", r))
		os.Exit(1)
	})

	roomCollector.OnHTML("div#divTT", func(tableContainerElement *colly.HTMLElement) {
		// https://github.com/piotrpdev/WIT-Timetable-Generator/blob/main/generateJson.js
		// TODO: Check table was found for each room request sent
		slog.Debug("Found table container")

		// TODO: Sanity check the column titles are as expected

		timetable := Timetable{Error: false, Empty: true, FreeTimes: make(map[string][]string)}

		tableContainerElement.ForEach("table:nth-child(1)", func(_ int, headerElement *colly.HTMLElement) {
			slog.Debug("Found header")

			date := headerElement.ChildText("tbody > tr:nth-child(1) > td[align='Right'] > b")

			if date != "" {
				splitDate := strings.Split(date, " ")
				timetable.Date = splitDate[1]
			} else {
				slog.Error("Header date is empty")
				timetable.Error = true
			}

			room := headerElement.ChildText("tbody > tr:nth-child(3) > td[align='Center'] > b")

			if room != "" {
				splitRoom := strings.Split(room, " ")
				timetable.Room = splitRoom[2]
			} else {
				slog.Error("Header room is empty")
				timetable.Error = true
			}

			week := headerElement.ChildText("tbody > tr:nth-child(3) > td[align='Right'] > b")

			if week != "" {
				splitWeek := strings.Split(week, " ")
				timetable.Week = splitWeek[2]
			} else {
				slog.Error("Header week is empty")
				timetable.Error = true
			}
		})

		tableContainerElement.ForEach("table:nth-child(2)", func(_ int, timetableElement *colly.HTMLElement) {
			slog.Info("Found timetable, parsing...")

			currentDay := -1

			timetableElement.ForEach("tr:not(:first-child)", func(_ int, timetableRowElement *colly.HTMLElement) {
				slog.Debug("Found row")

				dayName := timetableRowElement.ChildText("td[colspan='11'] > strong > font > i")

				if slices.Contains(days, dayName) {
					slog.Debug("(Skip) Row contains day", slog.String("dayName", dayName))
					currentDay++
					return
				}

				subject := timetableRowElement.ChildText("td:nth-of-type(5)")

				if subject != "" {
					slog.Debug("(Skip) Row contains subject", slog.String("subject", subject))
					timetable.Empty = false
					return
				}

				time := timetableRowElement.ChildText("td:nth-of-type(1)")

				switch currentDay {
				case -1:
					slog.Error("'currentDay' switch got '-1'")
					timetable.Error = true
				default:
					timetable.FreeTimes[days[currentDay]] = append(timetable.FreeTimes[days[currentDay]], time)
				}
			})
		})

		timetableList = append(timetableList, timetable)
	})

	roomCollector.OnRequest(func(r *colly.Request) {
		data, _ := io.ReadAll(r.Body)
		stringData := string(data)
		queryValues, _ := url.ParseQuery(stringData)

		slog.Info("Requesting room", slog.String("room", queryValues.Get("CboLocation")))
	})

	roomCollector.OnError(func(r *colly.Response, err error) {
		slog.Error(err.Error())
		slog.Error("Requesting room failed", slog.String("URL", r.Request.URL.String()), slog.Any("response", r))
		os.Exit(1)
	})

	initCollector.Visit(roomTimetableUrl)

	initCollector.Wait()
	roomCollector.Wait()

	errorCount := 0
	emptyTimetableCount := 0
	freeTimesCount := 0
	timetableCount := 0
	// [Monday, Tuesday] - [17:15, 18:15] - [IT101, IT102]
	freeRoomTable := make(map[string]map[string][]string)

	for _, timetable := range timetableList {
		timetableCount++

		if timetable.Error {
			errorCount++
		}

		if timetable.Empty {
			slog.Debug("Found empty timetable", slog.String("room", timetable.Room))
			emptyTimetableCount++
		}

		for _, day := range days {
			for _, time := range timetable.FreeTimes[day] {
				freeTimesCount++
				if freeRoomTable[day] == nil {
					freeRoomTable[day] = map[string][]string{}
				}
				freeRoomTable[day][time] = append(freeRoomTable[day][time], timetable.Room)
			}
		}
	}

	slog.Info(
		"Finished all requests", slog.Int("timetableCount", timetableCount),
		slog.Int("errors", errorCount),
		slog.Int("emptyTimetables", emptyTimetableCount),
		slog.Int("freeTimesCount", freeTimesCount))

	// TODO: Print nice ASCII table to stdout (and file) <https://github.com/olekukonko/tablewriter>
	slog.Info("Printing JSON Marshalled freeRoomTable")
	marshalledJson, err := json.Marshal(freeRoomTable)
	if err != nil {
		slog.Error(err.Error())
		slog.Error("Failed to JSON Marshal freeRoomTable")
		os.Exit(1)
	} else {
		slog.Info("JSON Marshal worked", slog.String("marshalledJson", string(marshalledJson)))
	}

	currentTime := getTimeStamp()

	slog.Info("Printing JSON MarshalIndented freeRoomTable to stdout and saving to file")
	marshalledIndentedJson, err := json.MarshalIndent(freeRoomTable, "", "  ")
	if err != nil {
		slog.Error(err.Error())
		slog.Error("Failed to JSON MarshalIndent freeRoomTable")
		os.Exit(1)
	} else {
		fmt.Println(string(marshalledIndentedJson))
		os.WriteFile(fmt.Sprintf("./%s_freeRoomTable.json", currentTime), marshalledIndentedJson, 0644)
	}

	slog.Info("Attempting to create ASCII table of freeRoomTable")

	var asciiTableBuffer strings.Builder
	var multiTableWriter io.Writer
	asciiFilename := fmt.Sprintf("./%s_ascii.txt", currentTime)

	asciiTableFile, err := os.OpenFile(asciiFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		slog.Error(err.Error())
		slog.Error("Failed to create ASCII table file, just using stdout", slog.String("failedFile", asciiFilename))
		multiTableWriter = io.MultiWriter(os.Stdout, &asciiTableBuffer)
	} else {
		multiTableWriter = io.MultiWriter(os.Stdout, &asciiTableBuffer, asciiTableFile)
	}

	asciiTable := tablewriter.NewWriter(multiTableWriter)
	asciiTable.SetHeader(append([]string{""}, days...))
	asciiTable.SetRowLine(true)

	asciiTableData := [][]string{}

	for _, time := range times {
		asciiTableData = append(asciiTableData, []string{time, "", "", "", "", ""})
	}

	slog.Info("Attempting to create rows of ASCII table")
	for dayIdx, day := range days {
		for freeTime, rooms := range freeRoomTable[day] {
			indexOfTime := SliceIndex(len(times), func(i int) bool { return times[i] == freeTime })
			if asciiTableData[indexOfTime][1+dayIdx] == "" {
				asciiTableData[indexOfTime][1+dayIdx] += strings.Join(rooms, ", ")
			} else {
				asciiTableData[indexOfTime][1+dayIdx] += (", " + strings.Join(rooms, ", "))
			}
		}
	}

	for _, row := range asciiTableData {
		asciiTable.Append(row)
	}

	slog.Info("Rendering ASCII table to stdout")
	asciiTable.Render()

	// webhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")

	// if len(webhookUrl) > 0 {
	// 	sendAsciiFileToWebhook(asciiTableBuffer.String(), webhookUrl)
	// }
}
