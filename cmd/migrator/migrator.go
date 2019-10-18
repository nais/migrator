package main

import flag "github.com/spf13/pflag"

type Config struct {
	Fasit         bool
	FasitUser     string
	FasitPassword string
	FasitUrl      string
}

var (
	cfg = Config{
		FasitUrl: "http://localhost:8080",
	}
)

func init() {
	flag.BoolVar(&cfg.Fasit, "fasit", cfg.Fasit, "use fasit integration")
	flag.StringVar(&cfg.FasitUser, "fasit-user", cfg.FasitUser, "fasit username")
	flag.StringVar(&cfg.FasitPassword, "fasit-password", cfg.FasitPassword, "fasit password")
}

func main() {
	flag.Parse()
}
