    // Execute the SQL query
    rows, err := db.Query(sqlQuery)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    // Fetch the column names
    columns, err := rows.Columns()
    if err != nil {
        log.Fatal(err)
    }

    // Prepare to scan the data
    values := make([]interface{}, len(columns))
    valuePtrs := make([]interface{}, len(columns))
    for i := range columns {
        valuePtrs[i] = &values[i]
    }

    // Store the result in a slice of strings
    var result []string

    // Iterate over rows and append them to the result slice
    for rows.Next() {
        err := rows.Scan(valuePtrs...)
        if err != nil {
            log.Fatal(err)
        }

        var line []string
        for _, val := range values {
            var str string
            switch v := val.(type) {
            case nil:
                str = "NULL"
            case []byte:
                str = string(v)
            default:
                str = fmt.Sprintf("%v", v)
            }
            line = append(line, str)
        }
        result = append(result, strings.Join(line, ", "))
    }

    // Print the result
    fmt.Println("Query Result:")
    for _, line := range result {
        fmt.Println(line)
    }

    // Run OS commands like sort and uniq
    if len(result) > 0 {
        cmd := exec.Command("sh", "-c", "echo \""+strings.Join(result, "\n")+"\" | sort | uniq")
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

        err := cmd.Run()
        if err != nil {
            log.Fatal(err)
        }
    }
