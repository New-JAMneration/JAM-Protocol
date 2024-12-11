- **/cmd/** 儲存應用程式的主要可執行檔，通常包含專案的進入點子模組。
- **/pkg/** 放置通用的代碼庫（Library），供其他應用或模組重複使用。
- **/internal/** 包含內部的專案模組，這些模組不應被外部的應用程序或模組依賴。
- **/logger/** 全局印log
    ```golang
        // Debug log
        logger.Debug("This is a debug message", "DEBUG-001")
    
        // Info log
        logger.Info("This is an info message", "INFO-001")
    
        // Warning log
        logger.Warn("This is a warning message", "WARN-001")
    
        // Error log
        logger.Error("This is an error message", "ERROR-001")
    ```