package igs

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProductType represents the type of IGS product
type ProductType string

const (
	// ProductTypeSP3 represents precise orbit files (.sp3)
	ProductTypeSP3 ProductType = "sp3"
	// ProductTypeCLK represents precise clock files (.clk)
	ProductTypeCLK ProductType = "clk"
)

// AnalysisCenter represents an IGS analysis center
type AnalysisCenter string

const (
	// AnalysisCenterIGS represents the International GNSS Service
	AnalysisCenterIGS AnalysisCenter = "igs"
	// AnalysisCenterCOD represents the Center for Orbit Determination in Europe
	AnalysisCenterCOD AnalysisCenter = "cod"
	// AnalysisCenterEMR represents Natural Resources Canada
	AnalysisCenterEMR AnalysisCenter = "emr"
	// AnalysisCenterESA represents the European Space Agency
	AnalysisCenterESA AnalysisCenter = "esa"
	// AnalysisCenterGFZ represents GeoForschungsZentrum Potsdam
	AnalysisCenterGFZ AnalysisCenter = "gfz"
	// AnalysisCenterJPL represents Jet Propulsion Laboratory
	AnalysisCenterJPL AnalysisCenter = "jpl"
)

// Client represents an IGS products client
type Client struct {
	// BaseURL is the base URL for the IGS products
	BaseURL string
	// HTTPClient is the HTTP client used for requests
	HTTPClient *http.Client
	// DownloadDir is the directory where files will be downloaded
	DownloadDir string
}

// NewClient creates a new IGS client
func NewClient(downloadDir string) *Client {
	return &Client{
		BaseURL:     "https://igs.ign.fr/pub/igs/products/",
		HTTPClient:  &http.Client{Timeout: 60 * time.Second},
		DownloadDir: downloadDir,
	}
}

// SetBaseURL sets the base URL for the IGS products
func (c *Client) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
}

// GPSWeekAndDay calculates the GPS week and day of week from a time
func GPSWeekAndDay(t time.Time) (int, int) {
	// GPS time started at 00:00:00 January 6, 1980 UTC
	gpsStartTime := time.Date(1980, 1, 6, 0, 0, 0, 0, time.UTC)

	// Calculate the duration since the GPS start time
	duration := t.Sub(gpsStartTime)

	// Calculate the number of days since the GPS start time
	days := int(duration.Hours() / 24)

	// Calculate the GPS week
	week := days / 7

	// Calculate the day of the week (0-6, where 0 is Sunday)
	dayOfWeek := days % 7

	return week, dayOfWeek
}

// GetProductURL generates the URL for a specific product
func (c *Client) GetProductURL(t time.Time, productType ProductType, ac AnalysisCenter) (string, error) {
	week, day := GPSWeekAndDay(t)

	// Format: {ac}{week}{day}.{ext}
	// Example: igs20607.sp3.Z for IGS orbit file for GPS week 2060, day 7
	filename := fmt.Sprintf("%s%04d%d.%s", ac, week, day, productType)

	// Add compression extension if needed
	if productType == ProductTypeSP3 || productType == ProductTypeCLK {
		filename += ".Z"
	}

	// Construct the full URL
	url := fmt.Sprintf("%s/%04d/%s", c.BaseURL, week, filename)

	return url, nil
}

// DownloadFile downloads a file from the given URL to the specified local path
func (c *Client) DownloadFile(url, localPath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Write the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DownloadProduct downloads a specific product for the given time
func (c *Client) DownloadProduct(t time.Time, productType ProductType, ac AnalysisCenter) (string, error) {
	// Get the URL for the product
	url, err := c.GetProductURL(t, productType, ac)
	if err != nil {
		return "", fmt.Errorf("failed to get product URL: %w", err)
	}

	// Generate the local file path
	week, day := GPSWeekAndDay(t)
	filename := fmt.Sprintf("%s%04d%d.%s.Z", ac, week, day, productType)
	localPath := filepath.Join(c.DownloadDir, fmt.Sprintf("%04d", week), filename)

	// Download the file
	err = c.DownloadFile(url, localPath)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}

	return localPath, nil
}

// DownloadSP3 downloads a SP3 file for the given time
func (c *Client) DownloadSP3(t time.Time, ac AnalysisCenter) (string, error) {
	return c.DownloadProduct(t, ProductTypeSP3, ac)
}

// DownloadCLK downloads a CLK file for the given time
func (c *Client) DownloadCLK(t time.Time, ac AnalysisCenter) (string, error) {
	return c.DownloadProduct(t, ProductTypeCLK, ac)
}

// DecompressFile decompresses a .Z file using the uncompress command
// This assumes that the uncompress command is available on the system
func DecompressFile(filePath string) (string, error) {
	if !strings.HasSuffix(filePath, ".Z") {
		return "", errors.New("file is not compressed with .Z")
	}

	// Get the output file path (remove the .Z extension)
	outputPath := strings.TrimSuffix(filePath, ".Z")

	// Use the uncompress command to decompress the file
	cmd := fmt.Sprintf("uncompress -f %s", filePath)

	// Execute the command
	err := exec.Command("sh", "-c", cmd).Run()
	if err != nil {
		return "", fmt.Errorf("failed to decompress file: %w", err)
	}

	return outputPath, nil
}
