// Package stream provides stream input/output functionality for GNSS data
package stream

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bramburn/gnssgo/pkg/gnssgo/gtime"
)

// File constants
const (
	tagExt       = ".tag"    // Time tag file extension
	TIMETAG      = "TIMETAG" // Time tag header
	FILETAGH_LEN = 4         // File tag header length
)

// OpenStreamFile opens a file stream
// path format: filepath[::T[::+<off>][::x<speed>]][::S=swapintv][::P={4|8}]
func OpenStreamFile(path string, mode int, msg *string) *FileType {
	var (
		file                   *FileType = new(FileType)
		time0                  gtime.Gtime
		speed, start, swapintv float64 = 1.0, 0.0, 0.0
		timetag, size_fpos     int     = 0, 4 /* default 4B */
	)

	Tracet(3, "OpenStreamFile: path=%s mode=%d\n", path, mode)

	if mode&(STR_MODE_R|STR_MODE_W) == 0 {
		return nil
	}

	// Parse path options
	filePath := path
	options := strings.Split(path, "::")
	if len(options) > 1 {
		filePath = options[0]

		for i := 1; i < len(options); i++ {
			opt := options[i]

			if opt == "T" {
				timetag = 1
			} else if strings.HasPrefix(opt, "+") {
				start, _ = strconv.ParseFloat(opt[1:], 64)
			} else if strings.HasPrefix(opt, "x") {
				speed, _ = strconv.ParseFloat(opt[1:], 64)
			} else if strings.HasPrefix(opt, "S=") {
				swapintv, _ = strconv.ParseFloat(opt[2:], 64)
			} else if strings.HasPrefix(opt, "P=") {
				size_fpos, _ = strconv.Atoi(opt[2:])
			}
		}
	}

	// Validate parameters
	if start < 0.0 {
		start = 0.0
	}
	if swapintv < 0.0 {
		swapintv = 0.0
	}
	if size_fpos != 8 {
		size_fpos = 4
	}

	// Initialize file structure
	file.path = filePath
	file.openpath = ""
	file.mode = mode
	file.timetag = timetag
	file.repmode = 0
	file.offset = 0
	file.size_fpos = size_fpos
	file.time = time0
	file.wtime = time0
	file.tick = 0
	file.tick_f = 0
	file.fpos_n = 0
	file.tick_n = 0
	file.start = start
	file.speed = speed
	file.swapintv = swapintv

	// Get current time
	time := gtime.Utc2GpsT(gtime.TimeGet())

	// Open new file
	if openfile_(file, time, msg) == 0 {
		return nil
	}

	return file
}

// openfile_ opens a file with time tag
func openfile_(file *FileType, time gtime.Gtime, msg *string) int {
	var (
		tagpath string
		tagh    []byte = make([]byte, TIMETAGH_LEN)
		err     error
	)

	Tracet(3, "openfile_: path=%s time=%s\n", file.path, gtime.TimeStr(time, 0))

	file.time = gtime.Utc2GpsT(gtime.TimeGet())
	file.tick = TickGet()
	file.tick_f = file.tick
	file.fpos_n = 0
	file.tick_n = 0

	// Use stdin or stdout if file path is null
	if len(file.path) == 0 {
		if file.mode&STR_MODE_R > 0 {
			file.fp = os.Stdin
		} else {
			file.fp = os.Stdout
		}
		return 1
	}

	// Replace file path keywords
	rpath := reppath(file.path, time, "", "")

	// Open specific file based on mode
	if file.mode&STR_MODE_R > 0 { // Read mode
		file.fp, err = os.Open(rpath)
		if err != nil {
			if msg != nil {
				*msg = fmt.Sprintf("file open error: %s", err.Error())
			}
			Tracet(2, "openfile: file open error: %s\n", err.Error())
			return 0
		}

		// Open tag file if needed
		if file.timetag > 0 {
			tagpath = rpath + tagExt
			file.fp_tag, err = os.Open(tagpath)
			if err != nil {
				if msg != nil {
					*msg = fmt.Sprintf("tag file open error: %s", err.Error())
				}
				Tracet(2, "openfile: tag file open error: %s\n", err.Error())
				file.fp.Close()
				return 0
			}

			// Read time tag header
			if _, err := file.fp_tag.Read(tagh); err != nil || string(tagh[:TIMETAGH_LEN-1]) != TIMETAG {
				if msg != nil {
					*msg = "tag file format error"
				}
				Tracet(2, "openfile: tag file format error\n")
				file.fp.Close()
				file.fp_tag.Close()
				return 0
			}
		}
	} else { // Write mode
		// Create directory if it doesn't exist
		dir := filepath.Dir(rpath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				if msg != nil {
					*msg = fmt.Sprintf("directory creation error: %s", err.Error())
				}
				Tracet(2, "openfile: directory creation error: %s\n", err.Error())
				return 0
			}
		}

		// Open file for writing
		file.fp, err = os.Create(rpath)
		if err != nil {
			if msg != nil {
				*msg = fmt.Sprintf("file open error: %s", err.Error())
			}
			Tracet(2, "openfile: file open error: %s\n", err.Error())
			return 0
		}

		// Open tag file if needed
		if file.timetag > 0 {
			tagpath = rpath + tagExt
			file.fp_tag, err = os.Create(tagpath)
			if err != nil {
				if msg != nil {
					*msg = fmt.Sprintf("tag file open error: %s", err.Error())
				}
				Tracet(2, "openfile: tag file open error: %s\n", err.Error())
				file.fp.Close()
				return 0
			}

			// Write time tag header
			copy(tagh[:TIMETAGH_LEN-1], []byte(TIMETAG))
			tagh[TIMETAGH_LEN-1] = byte(file.size_fpos)
			if _, err := file.fp_tag.Write(tagh); err != nil {
				if msg != nil {
					*msg = fmt.Sprintf("tag file write error: %s", err.Error())
				}
				Tracet(2, "openfile: tag file write error: %s\n", err.Error())
				file.fp.Close()
				file.fp_tag.Close()
				return 0
			}
		}
	}

	file.openpath = rpath
	return 1
}

// CloseFile closes a file stream
func (file *FileType) CloseFile() {
	Tracet(3, "CloseFile:\n")

	if file == nil {
		return
	}

	closefile_(file)
}

// closefile_ closes a file
func closefile_(file *FileType) {
	Tracet(3, "closefile_: path=%s\n", file.path)

	if file.fp != nil && file.fp != os.Stdin && file.fp != os.Stdout {
		file.fp.Close()
		file.fp = nil
	}

	if file.fp_tag != nil {
		file.fp_tag.Close()
		file.fp_tag = nil
	}

	if file.fp_tmp != nil {
		file.fp_tmp.Close()
		file.fp_tmp = nil
	}

	if file.fp_tag_tmp != nil {
		file.fp_tag_tmp.Close()
		file.fp_tag_tmp = nil
	}
}

// ReadFile reads data from a file stream
func (file *FileType) ReadFile(buff []byte, n int64, msg *string) int {
	// No variables needed for simplified implementation

	Tracet(4, "ReadFile: n=%d\n", n)

	if file == nil || file.fp == nil {
		return 0
	}

	// Lock file for thread safety
	file.lock.Lock()
	defer file.lock.Unlock()

	// If not time-tagged file, simply read the data
	if file.timetag <= 0 {
		nr, err := file.fp.Read(buff[:n])
		if err != nil {
			if msg != nil {
				*msg = fmt.Sprintf("file read error: %s", err.Error())
			}
			return 0
		}
		return nr
	}

	// For simplicity, we'll just read the data without time tag handling for now
	nr, err := file.fp.Read(buff[:n])
	if err != nil {
		if msg != nil {
			*msg = fmt.Sprintf("file read error: %s", err.Error())
		}
		return 0
	}
	return nr
}

// readfiletag reads file position and tick from tag file (placeholder)
func readfiletag(file *FileType, fpos *int64, tick *uint32, msg *string) int {
	// Simplified implementation
	return 1
}

// readfiletime reads data from file for a specific time (placeholder)
func readfiletime(file *FileType, buff []byte, n int64, tow float64, msg *string) int {
	// Simplified implementation
	nr, err := file.fp.Read(buff[:n])
	if err != nil {
		*msg = fmt.Sprintf("file read error: %s", err.Error())
		return -1
	}

	return nr
}

// WriteFile writes data to a file stream
func (file *FileType) WriteFile(buff []byte, n int, msg *string) int {
	var (
		wtime gtime.Gtime
		tagb  []byte
		tow   float64
		week  int
	)

	Tracet(4, "WriteFile: n=%d\n", n)

	if file == nil || file.fp == nil {
		return 0
	}

	// Lock file for thread safety
	file.lock.Lock()
	defer file.lock.Unlock()

	// Write data to file
	nw, err := file.fp.Write(buff[:n])
	if err != nil {
		if msg != nil {
			*msg = fmt.Sprintf("file write error: %s", err.Error())
		}
		return 0
	}

	// Write time tag if enabled
	if file.timetag > 0 && file.fp_tag != nil {
		// Get current file position
		fpos, err := file.fp.Seek(0, 1) // 1 = SEEK_CUR
		if err != nil {
			if msg != nil {
				*msg = fmt.Sprintf("file seek error: %s", err.Error())
			}
			return 0
		}

		// Get current tick
		tick := TickGet()

		// Allocate tag buffer based on file position size
		if file.size_fpos == 8 {
			tagb = make([]byte, FILETAGH_LEN+8)
			tagb[0] = byte(fpos)
			tagb[1] = byte(fpos >> 8)
			tagb[2] = byte(fpos >> 16)
			tagb[3] = byte(fpos >> 24)
			tagb[4] = byte(fpos >> 32)
			tagb[5] = byte(fpos >> 40)
			tagb[6] = byte(fpos >> 48)
			tagb[7] = byte(fpos >> 56)
			tagb[8] = byte(tick)
			tagb[9] = byte(tick >> 8)
			tagb[10] = byte(tick >> 16)
			tagb[11] = byte(tick >> 24)
		} else {
			tagb = make([]byte, FILETAGH_LEN+4)
			tagb[0] = byte(fpos)
			tagb[1] = byte(fpos >> 8)
			tagb[2] = byte(fpos >> 16)
			tagb[3] = byte(fpos >> 24)
			tagb[4] = byte(tick)
			tagb[5] = byte(tick >> 8)
			tagb[6] = byte(tick >> 16)
			tagb[7] = byte(tick >> 24)
		}

		// Write tag data
		if _, err := file.fp_tag.Write(tagb); err != nil {
			if msg != nil {
				*msg = fmt.Sprintf("tag file write error: %s", err.Error())
			}
			return 0
		}
	}

	// Swap files if needed
	if file.swapintv > 0.0 && file.wtime.Time != 0 {
		wtime = gtime.Utc2GpsT(gtime.TimeGet())
		tow = gtime.Time2GpsT(file.wtime, &week)
		tow = gtime.Time2GpsT(wtime, &week) - tow

		if tow < 0.0 {
			tow += 604800.0
		}

		if tow > file.swapintv*3600.0 {
			swapfile(file)
		}
	}

	return nw
}

// swapfile swaps files for data and tag
func swapfile(file *FileType) {
	var (
		time gtime.Gtime
		msg  string
	)

	Tracet(3, "swapfile:\n")

	// Close current files
	closefile_(file)

	// Get current time
	time = gtime.Utc2GpsT(gtime.TimeGet())

	// Open new files
	openfile_(file, time, &msg)

	// Update write time
	file.wtime = time
}

// reppath replaces file path keywords with date/time and other information
func reppath(path string, time gtime.Gtime, sta, ext string) string {
	// This is a placeholder for the reppath implementation
	// The actual implementation would replace keywords like %Y, %m, %d, etc.
	// with the corresponding date/time values

	// For now, just return the original path
	return path
}

// StateXFile returns the state of a file stream
func (file *FileType) StateXFile(msg *string) int {
	return 0
}
