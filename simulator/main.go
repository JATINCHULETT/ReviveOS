package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/reviveos/simulator/generator"
)

func main() {
	count := flag.Int("count", 500, "Number of events to generate")
	output := flag.String("output", "dataset.json", "Output file for the generated dataset")
	flag.Parse()

	gen := generator.New()
	fmt.Printf("Generating %d synthetic payment failure events...\n", *count)

	events := gen.GenerateBatch(*count)

	file, err := os.Create(*output)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(events); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}

	fmt.Printf("Successfully wrote dataset to %s\n", *output)
}
