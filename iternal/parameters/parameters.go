package parameters

import (
	"flag"
	"os"
)

type Parameters struct {
	RunAddr           string
	DataBaseURI       string
	AccuralSystemAddr string
}

func ParseFlags() (p Parameters) {
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.StringVar(&p.RunAddr, "a", "localhost:8081", "address and port to run server")
	f.StringVar(&p.DataBaseURI, "d", "", "connection string to database")
	f.StringVar(&p.AccuralSystemAddr, "r", "localhost:8080", "address and port to accural system")
	f.Parse(os.Args[1:])

	if envAddr := os.Getenv("RUN_ADDRESS"); envAddr != "" {
		p.RunAddr = envAddr
	}

	if envDB := os.Getenv("DATABASE_URI"); envDB != "" {
		p.DataBaseURI = envDB
	}

	if envAS := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAS != "" {
		p.AccuralSystemAddr = envAS
	}

	return
}
