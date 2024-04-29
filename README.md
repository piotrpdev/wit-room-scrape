# WIT Room Scrape

Finds when rooms in the SETU Waterford campus are empty.

Saves output to an `ASCII` table file and `JSON` file.

## Example

### Finding empty IT rooms

#### Regexp

```go
roomRegexp, _ := regexp.Compile("^IT.{3}$")
```

#### Output

```txt
+-------+--------------------------------+--------------------------------+--------------------------------+--------------------------------+--------------------------------+
|       |             MONDAY             |            TUESDAY             |           WEDNESDAY            |            THURSDAY            |             FRIDAY             |
+-------+--------------------------------+--------------------------------+--------------------------------+--------------------------------+--------------------------------+
| 09:15 | IT102, IT112, IT119, IT220,    | IT102, IT112, IT119, IT221     | IT112, IT119, IT201, IT220,    | IT102, IT112, IT201            | IT102, IT112, IT201, IT221,    |
|       | IT222, ITG02, ITG17            |                                | ITG03, ITG17                   |                                | IT222, ITG02                   |
| 10:15 | IT102, IT112, IT119, ITG17     | IT102, IT112, IT222, ITG18     | IT112, IT119, ITG17            | IT102, IT112                   | IT102, IT112, IT119, IT201,    |
|       |                                |                                |                                |                                | IT222, ITG18                   |
| 11:15 | IT102, IT112, IT120, IT201,    | IT103, IT112, ITG17            | IT112, IT119, ITG18            | IT102, IT112, IT120, IT203,    | IT102, IT112                   |
|       | IT203, IT220, IT221, ITG01,    |                                |                                | IT220, IT221                   |                                |
|       | ITG02, ITG18, ITG19            |                                |                                |                                |                                |
| 12:15 | IT102, IT112, IT118, IT201,    | IT103, IT112, ITG17, ITG18     | IT112, IT118, IT201, ITG01,    | IT102, IT103, IT112, ITG02,    | IT102, IT112                   |
|       | IT220, IT221, ITG01, ITG17,    |                                | ITG17, ITG18, ITG19            | ITG19                          |                                |
|       | ITG18, ITG19                   |                                |                                |                                |                                |
| 13:15 | IT101, IT102, IT112, IT118,    | IT102, IT103, IT112, ITG03     | IT112, IT202, ITG03            | IT112, IT118, IT119, ITG03,    | IT101, IT102, IT112, IT118,    |
|       | IT119, IT120, IT201, IT202,    |                                |                                | ITG18, ITG19                   | IT120, IT201, IT202, IT220,    |
|       | IT221, ITG01, ITG03, ITG17,    |                                |                                |                                | ITG01, ITG17, ITG18, ITG19     |
|       | ITG19                          |                                |                                |                                |                                |
| 14:15 | IT101, IT102, IT112, IT120,    | IT103, IT112, IT220            | IT112, ITG02                   | IT112, IT118, ITG17            | IT101, IT112, IT120, IT202,    |
|       | IT201, IT221, ITG01, ITG02,    |                                |                                |                                | IT222, ITG01, ITG03, ITG18,    |
|       | ITG18                          |                                |                                |                                | ITG19                          |
| 15:15 | IT103, IT112, ITG02, ITG18     | IT112, IT220, ITG01, ITG02,    | IT103, IT112, ITG02, ITG03,    | IT102, IT112                   | IT101, IT112, IT119, IT120,    |
|       |                                | ITG18                          | ITG18                          |                                | IT202, IT221, IT222, ITG01,    |
|       |                                |                                |                                |                                | ITG03, ITG17, ITG18, ITG19     |
| 16:15 | IT103, IT112, IT119, IT120,    | IT112, IT220, ITG01, ITG02,    | IT103, IT112, IT202, ITG02,    | IT102, IT103, IT112            | IT101, IT102, IT103, IT112,    |
|       | ITG02, ITG03, ITG18, ITG19     | ITG18                          | ITG03                          |                                | IT118, IT119, IT120, IT201,    |
|       |                                |                                |                                |                                | IT202, IT220, IT221, IT222,    |
|       |                                |                                |                                |                                | ITG01, ITG02, ITG03, ITG17,    |
|       |                                |                                |                                |                                | ITG18, ITG19                   |
| 17:15 | IT101, IT102, IT103, IT112,    | IT101, IT102, IT103, IT112,    | IT101, IT102, IT103, IT112,    | IT101, IT102, IT103, IT112,    | IT101, IT102, IT103, IT112,    |
|       | IT118, IT119, IT120, IT201,    | IT118, IT119, IT120, IT201,    | IT118, IT119, IT120, IT201,    | IT118, IT119, IT120, IT201,    | IT118, IT119, IT120, IT201,    |
|       | IT202, IT203, IT220, IT221,    | IT202, IT203, IT220, IT221,    | IT202, IT203, IT220, IT221,    | IT202, IT203, IT220, IT221,    | IT202, IT203, IT220, IT221,    |
|       | IT222, ITG01, ITG02, ITG03,    | IT222, ITG01, ITG02, ITG03,    | IT222, ITG01, ITG02, ITG03,    | IT222, ITG01, ITG02, ITG03,    | IT222, ITG01, ITG02, ITG03,    |
|       | ITG17, ITG18, ITG19            | ITG17, ITG18, ITG19            | ITG17, ITG18, ITG19            | ITG17, ITG18, ITG19            | ITG17, ITG18, ITG19            |
+-------+--------------------------------+--------------------------------+--------------------------------+--------------------------------+--------------------------------+
```

## License

GNU GPLv3, see [LICENSE](./LICENSE).
