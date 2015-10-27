package scoring

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

type Submission struct {
	Pk            int
	CompetitionPk int
	Path          string
}

type PlainRC struct {
	filerc *os.File
}

func NewPlainRC(path string) (*PlainRC, error) {
	rc := PlainRC{filerc: nil}
	var err error
	rc.filerc, err = os.Open(path)
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (rc *PlainRC) Read(b []byte) (int, error) {
	return rc.filerc.Read(b)
}

func (rc *PlainRC) Close() error {
	return rc.filerc.Close()
}

type ZipRC struct {
	ziprc  *zip.ReadCloser
	filerc io.ReadCloser
}

func NewZipRC(path string) (*ZipRC, error) {
	rc := ZipRC{ziprc: nil, filerc: nil}
	var err error
	rc.ziprc, err = zip.OpenReader(path)
	if err != nil {
		return nil, errors.New("Fail to open zip file.")
	}

	var target *zip.File
	target = nil
	for _, f := range rc.ziprc.File {
		if strings.HasSuffix(f.Name, "csv") {
			if target != nil {
				return nil, errors.New("Multiple csv files exist.")
			}
			target = f
		}
	}
	if target == nil {
		return nil, errors.New("No csv files found.")
	}

	rc.filerc, err = target.Open()
	if err != nil {
		return nil, errors.New("Fail to open csv file.")
	}
	return &rc, nil
}
func (rc *ZipRC) Close() error {
	rc.filerc.Close()
	return rc.ziprc.Close()
}
func (rc *ZipRC) Read(b []byte) (int, error) {
	return rc.filerc.Read(b)
}

func (sub *Submission) Open() (io.ReadCloser, error) {
	if strings.HasSuffix(sub.Path, ".csv") {
		return NewPlainRC(sub.Path)
	} else if strings.HasSuffix(sub.Path, ".zip") {
		return NewZipRC(sub.Path)
	} else {
		return nil, errors.New("File type not supported")
	}
}

func (submission *Submission) ReadData() (map[string]float32, error) {
	rc, err := submission.Open()
	if err != nil {
		return nil, err
	}
	res := map[string]float32{}
	csvReader := csv.NewReader(rc)
	msg := ""
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			msg = "Format error"
			break
		}
		if len(record) != 2 {
			msg = "Wrong column numbers"
			break
		}
		key := record[0]
		pred, err := strconv.ParseFloat(record[1], 32)
		if err != nil {
			msg = "Format error"
			break
		}
		if pred > 1 || pred < 0 {
			msg = "Prediction out of range"
			break
		}
		res[key] = float32(pred)
	}
	rc.Close()
	if msg != "" {
		return res, errors.New(msg)
	} else {
		return res, nil
	}
}