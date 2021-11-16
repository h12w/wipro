package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"h12.io/wipro/gen"
)

type Config struct {
	PackageName string
	OutputJSON  bool
	FilePrefix  string
}

func main() {
	var cfg Config
	flag.StringVar(&cfg.PackageName, "package", "proto", "package name")
	flag.BoolVar(&cfg.OutputJSON, "json", false, "output JSON")
	flag.StringVar(&cfg.FilePrefix, "prefix", "", "filename preifx")
	flag.Parse()

	if err := run(&cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg *Config) error {
	bnf := gen.DecodeBNF(os.Stdin)
	if cfg.OutputJSON {
		fmt.Println(bnf.JSON())
		fmt.Println(bnf.GoTypes().JSON())
		return nil
	}

	filePrefix := ""
	if cfg.FilePrefix != "" {
		filePrefix = cfg.FilePrefix + "_"
	}
	typesFilename := filePrefix + "types.go"
	funcsFilename := filePrefix + "funcs.go"

	typesFile, err := os.Create(typesFilename)
	if err != nil {
		return err
	}
	defer typesFile.Close()
	goTypes := bnf.GoTypes()
	goTypes.PackageName = cfg.PackageName
	if err := goTypes.Marshal(typesFile); err != nil {
		return err
	}

	funcsFile, err := os.Create(funcsFilename)
	if err != nil {
		return err
	}
	defer funcsFile.Close()
	goTypes.GoFuncs(funcsFile, cfg.PackageName)
	return nil
}
