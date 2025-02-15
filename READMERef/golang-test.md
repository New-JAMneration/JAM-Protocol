# Golang Test

```bash
-v 顯示詳細測試結果（verbose）
-cover 顯示測試覆蓋率
-failfast 發生錯誤就停止測試
-coverprofile 產生測試結果的檔案
-html 產生測試結果的 HTML 檔案
-short 跳過有 t.Skip 或 testing.Short() 的函式
```

### 測試某個 package

```bash
$ go test <package-name> -v
$ go test sandbox/go-sandbox/car -v
```

### 測試專案內的所有檔案

```bash
$ go test ./...
```

### 測試某資料夾內的所有檔案

```bash
$ go test ./car/...
```

### 只測試檔案中的某個 function

```bash
$ go test -run=TestCar_SetName -cover -v ./car/...
```

### 檢視測試覆蓋率

```bash
$ go test -cover .  # 只顯示在 Terminal
```

### 檢視測試報告及未被覆蓋到的程式碼

```bash
$ go test -coverprofile cover.out ./...
$ go tool cover -html=cover.out -o cover.html
$ open cover.html
```

### 清除測試的 cache

```bash
$ go clean -testcache
```

### 避免多個 package 的 test 同時執行

```bash
$ go test -p 1 ./...						# 限制 parallel 的數量為 1
```

### More details

For more details, please see [here](https://pjchender.dev/golang/note-golang-test/).

### 設定 TEST_MODE

預設 TEST_MODE 為 tiny

```bash
$ TEST_MODE=tiny go test ./internal/statistics
# or
$ TEST_MODE=full go test ./internal/statistics
```

> 切換測試時, 注意快取的問題, 可以透過 `go clean -testcache` 清除快取
