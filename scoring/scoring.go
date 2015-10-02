package scoring

import (
	"fmt"
	"strconv"
	"os"
	"io"
	"path"
	"log"
	"encoding/csv"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"database/sql"
	_ "github.com/lib/pq"
)


type Truth map[string]int;
type Competition struct {
	Public, Private Truth
}

type Submission struct {
	Pk int
	CompetitionPk int
	Path string
}

type Result struct {
	SubmissionPk int
	Message string
	PublicScore float32
	PrivateScore float32
}

type Daemon struct {
	config Config
	competitionTruths map[int]Competition	
	postgresConn *sql.DB
	redisConn redis.Conn
}

func (daemon *Daemon) connDb() {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", 
		daemon.config.PostgresUser, daemon.config.PostgresPassword, daemon.config.PostgresDb)

	fmt.Println(connStr)
	var err error
	daemon.postgresConn, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}
}

func (daemon *Daemon) connRedis() {
	var err error
	daemon.redisConn, err = redis.Dial("tcp", daemon.config.RedisHost+":"+strconv.Itoa(daemon.config.RedisPort))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func readTruth(path string) Truth{
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	reader := csv.NewReader(file)
	res := Truth{}
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		key := record[0]
		label := record[1]

		res[key], err = strconv.Atoi(label)
		if err != nil {
			log.Fatal(err)
		}
	}
	return res
}

func (daemon *Daemon) loadCompetitionData() {	
	rows, err := daemon.postgresConn.Query("SELECT id, public_truth, private_truth from competition_competition")
	if err != nil {
		log.Fatal(err)
	}

	daemon.competitionTruths = map[int]Competition{}
	for rows.Next() {
		var compId int
		var publicPath, privatePath string
		if err := rows.Scan(&compId, &publicPath, &privatePath); err != nil {
			log.Fatal(err)
		}

		daemon.competitionTruths[compId] = Competition {
			readTruth(path.Join(daemon.config.TruthRoot, publicPath)),
			readTruth(path.Join(daemon.config.TruthRoot, privatePath)),
		}
	}

	defer rows.Close()	
}

func (daemon *Daemon) Init(config Config){
	daemon.config = config
	daemon.connDb()
	daemon.connRedis()
	daemon.loadCompetitionData()
}

func (daemon *Daemon) Cleanup() {
	daemon.postgresConn.Close()
	daemon.redisConn.Close()
}

func (daemon *Daemon) Run(numWorker int){
	queue := make(chan Submission)
	for i:=0; i<numWorker; i+=1 {
		go daemon.work(queue)
	}

	for {
		submissionJson ,_ := redis.Bytes(daemon.redisConn.Do("LPOP", "submission_queue"))
		if submissionJson == nil {
			continue;
		}
		var submission Submission
		err := json.Unmarshal(submissionJson, &submission)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			queue <- submission
		}
	}
}
func Start(configFilename string, numWorker int){
	config, err := LoadConfig(configFilename)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	daemon := Daemon {}
	daemon.Init(config)
	defer daemon.Cleanup()

	daemon.Run(4)
}

func readSubmission(path string) (map[string]float32, string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	res := map[string]float32{}
	reader := csv.NewReader(file)
	msg := ""
	for {
		record, err := reader.Read()
		if err==io.EOF {
			break
		}
		if err!=nil {
			msg = "Format error" 
			break
		}
		if len(record) != 2{
			msg = "Wrong column numbers"
			break
		}
		key := record[0]
		pred, err := strconv.ParseFloat(record[1], 32)
		if err!=nil {
			msg = "Format error" 
			break
		}
		if pred>1 || pred<0 {
			msg = "Prediction out of range"
			break
		}
		res[key] = float32(pred)
	}

	return res, msg 
}

func evaluate(truth Truth, predication map[string]float32) float32{
	return AUC(truth, predication)
}

func (daemon *Daemon) writeMsg(submissionPk int, msg string) {
	log.Printf("Sub %d: %s\n", submissionPk, msg)
	_, err := daemon.postgresConn.Exec(`UPDATE competition_submission
							SET message=$1, status=3
							WHERE id=$2`, msg, submissionPk)
	if err != nil {
		log.Println(err.Error())
	}
}

func (daemon *Daemon) writeScore(submissionPk int, publicScore, privateScore float32) {
	log.Printf("Sub %d: %f %f\n", submissionPk, publicScore, privateScore)
	_, err := daemon.postgresConn.Exec(`UPDATE competition_submission
							SET public_score=$1, private_score=$2, status=2, message="Success"
							WHERE id=$3`, 
							publicScore, privateScore, submissionPk)
	if err != nil {
		log.Println(err.Error())
	}
}

func (daemon *Daemon) work(queue chan Submission){
	fmt.Println("worker")	

	for {
		submission := <- queue
		log.Printf("Get sub %d\n", submission.Pk)
		predication, msg := readSubmission(submission.Path)

		if msg != "" {
			daemon.writeMsg(submission.Pk, msg)
			continue
		}

		var publicScore, privateScore float32
		publicScore = evaluate(daemon.competitionTruths[submission.CompetitionPk].Public, predication)
		privateScore = evaluate(daemon.competitionTruths[submission.CompetitionPk].Private, predication)

		if msg != "" {
			daemon.writeMsg(submission.Pk, msg)
			continue
		}

		daemon.writeScore(submission.Pk, publicScore, privateScore)
	}
}