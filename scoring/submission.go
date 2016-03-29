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

type Prediction struct {
	classification_prediction map[string]float32
	rank_prediction           []rank
}

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

func (submission *Submission) ReadData(evaluation int) (Prediction, error) {
	rc, err := submission.Open()
	if err != nil {
		return Prediction{}, err
	}
	res := Prediction{}
	csvReader := csv.NewReader(rc)

	msg := ""
	if evaluation == 2 {
		res.classification_prediction = make(map[string]float32)
		
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
			res.classification_prediction[key] = float32(pred)
		}
	} else if evaluation == 1 {
		res.rank_prediction = make([]rank, 0)

		for {
			record, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				msg = "Format error"
				break
			}

			var line rank
			for j := 0; j < len(record); j++ {
				num, _ := strconv.Atoi(record[j])
				line = append(line, num)
			}
			res.rank_prediction = append(res.rank_prediction, line)
		}
	} else {
		msg = "evaluation method doesn't exist" 
	}
	
	rc.Close()
	if msg != "" {
		return res, errors.New(msg)
	} else {
		return res, nil
	}
}
