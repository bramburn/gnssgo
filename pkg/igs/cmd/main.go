package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bramburn/gnssgo/pkg/igs"
)

func main() {
	// Define command-line flags
	var (
		productType  string
		ac           string
		date         string
		outputDir    string
		decompress   bool
		listCenters  bool
		listProducts bool
	)

	flag.StringVar(&productType, "type", "sp3", "Product type (sp3 or clk)")
	flag.StringVar(&ac, "ac", "igs", "Analysis center (igs, cod, emr, esa, gfz, jpl)")
	flag.StringVar(&date, "date", "", "Date in YYYY-MM-DD format (default: today)")
	flag.StringVar(&outputDir, "out", ".", "Output directory")
	flag.BoolVar(&decompress, "decompress", false, "Decompress downloaded files")
	flag.BoolVar(&listCenters, "list-centers", false, "List available analysis centers")
	flag.BoolVar(&listProducts, "list-products", false, "List available product types")
	flag.Parse()

	// Handle listing options
	if listCenters {
		fmt.Println("Available analysis centers:")
		fmt.Println("  igs - International GNSS Service")
		fmt.Println("  cod - Center for Orbit Determination in Europe")
		fmt.Println("  emr - Natural Resources Canada")
		fmt.Println("  esa - European Space Agency")
		fmt.Println("  gfz - GeoForschungsZentrum Potsdam")
		fmt.Println("  jpl - Jet Propulsion Laboratory")
		return
	}

	if listProducts {
		fmt.Println("Available product types:")
		fmt.Println("  sp3 - Precise orbit files (.sp3)")
		fmt.Println("  clk - Precise clock files (.clk)")
		return
	}

	// Parse date
	var t time.Time
	var err error
	if date == "" {
		t = time.Now().UTC()
	} else {
		t, err = time.Parse("2006-01-02", date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing date: %v\n", err)
			os.Exit(1)
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Create client
	client := igs.NewClient(outputDir)

	// Map string to product type
	var pt igs.ProductType
	switch productType {
	case "sp3":
		pt = igs.ProductTypeSP3
	case "clk":
		pt = igs.ProductTypeCLK
	default:
		fmt.Fprintf(os.Stderr, "Invalid product type: %s\n", productType)
		os.Exit(1)
	}

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

	// Download the product
	fmt.Printf("Downloading %s product from %s for %s...\n", pt, analysisCenter, t.Format("2006-01-02"))
	filePath, err := client.DownloadProduct(t, pt, analysisCenter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading product: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloaded to: %s\n", filePath)

	// Decompress if requested
	if decompress {
		fmt.Println("Decompressing file...")
		decompressedPath, err := igs.DecompressFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decompressing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Decompressed to: %s\n", decompressedPath)
	}
}
