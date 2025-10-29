package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/mj3smile/bank-statement-processor/internal/handler"
)

func main() {
	http.HandleFunc("/statements", handler.UploadStatements)
	http.HandleFunc("/balance", handler.GetBalance)
	http.HandleFunc("transactions/issues", handler.GetTransactionIssues)

	http.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Hello World"))
	})

	err := http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
