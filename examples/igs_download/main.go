package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bramburn/gnssgo/pkg/igs"
)

func main() {
	// Parse command-line flags
	var (
		downloadSP3 bool
		downloadCLK bool
		ac          string
		date        string
		outputDir   string
		decompress  bool
	)

	flag.BoolVar(&downloadSP3, "sp3", true, "Download SP3 file")
	flag.BoolVar(&downloadCLK, "clk", true, "Download CLK file")
	flag.StringVar(&ac, "ac", "igs", "Analysis center (igs, cod, emr, esa, gfz, jpl)")
	flag.StringVar(&date, "date", "", "Date in YYYY-MM-DD format (default: today)")
	flag.StringVar(&outputDir, "out", "./data", "Output directory")
	flag.BoolVar(&decompress, "decompress", false, "Decompress downloaded files")
	flag.Parse()

	fmt.Println("GNSSGO IGS Product Download Example")
	fmt.Println("----------------------------------")

	// Parse date
	var t time.Time
	var err error
	if date == "" {
		t = time.Now().UTC()
		fmt.Printf("Using current date: %s\n", t.Format("2006-01-02"))
	} else {
		t, err = time.Parse("2006-01-02", date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing date: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Using specified date: %s\n", t.Format("2006-01-02"))
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create client
	client := igs.NewClient(outputDir)

	// Map string to analysis center
	var analysisCenter igs.AnalysisCenter
	switch ac {
	case "igs":
		analysisCenter = igs.AnalysisCenterIGS
	case "cod":
		analysisCenter = igs.AnalysisCenterCOD
	case "emr":
		analysisCenter = igs.AnalysisCenterEMR
	case "esa":
		analysisCenter = igs.AnalysisCenterESA
	case "gfz":
		analysisCenter = igs.AnalysisCenterGFZ
	case "jpl":
		analysisCenter = igs.AnalysisCenterJPL
	default:
		fmt.Fprintf(os.Stderr, "Invalid analysis center: %s\n", ac)
		os.Exit(1)
	}

	// Calculate GPS week and day
	week, day := igs.GPSWeekAndDay(t)
	fmt.Printf("GPS Week: %d, Day: %d\n", week, day)

	// Download SP3 file if requested
	if downloadSP3 {
		fmt.Printf("\nDownloading SP3 product from %s...\n", analysisCenter)
		filePath, err := client.DownloadSP3(t, analysisCenter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading SP3 product: %v\n", err)
		} else {
			fmt.Printf("Downloaded to: %s\n", filePath)

			// Decompress if requested
			if decompress {
				fmt.Println("Decompressing SP3 file...")
				decompressedPath, err := igs.DecompressFile(filePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error decompressing SP3 file: %v\n", err)
				} else {
					fmt.Printf("Decompressed to: %s\n", decompressedPath)
				}
			}
		}
	}

	// Download CLK file if requested
	if downloadCLK {
		fmt.Printf("\nDownloading CLK product from %s...\n", analysisCenter)
		filePath, err := client.DownloadCLK(t, analysisCenter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading CLK product: %v\n", err)
		} else {
			fmt.Printf("Downloaded to: %s\n", filePath)

			// Decompress if requested
			if decompress {
				fmt.Println("Decompressing CLK file...")
				decompressedPath, err := igs.DecompressFile(filePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error decompressing CLK file: %v\n", err)
				} else {
					fmt.Printf("Decompressed to: %s\n", decompressedPath)
				}
			}
		}
	}

	fmt.Println("\nDownload complete!")
}
