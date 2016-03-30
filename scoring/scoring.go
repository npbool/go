package scoring

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	_ "github.com/lib/pq"
)

type rank []int

type Truth struct {
	classification_truth map[string]int
	rank_truth           []rank
}

type Competition struct {
	Public, Private Truth
	NumLine         int
	evaluation		int
}

type Result struct {
	SubmissionPk int
	Message      string
	PublicScore  float32
	PrivateScore float32
}

type Daemon struct {
	config            Config
	competitionTruths map[int]Competition
	postgresConn      *sql.DB
	redisConn         redis.Conn
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

func readTruth(path string, evaluation int) Truth {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	res := Truth{}
	reader := csv.NewReader(file)
	

	if evaluation == 2 {
		res.classification_truth = make(map[string]int)

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

			res.classification_truth[key], err = strconv.Atoi(label)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else if evaluation == 1 {
		res.rank_truth = make([]rank, 0)

		for {
			record, err := reader.Read()

			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			var line rank
			for j := 0; j < len(record); j++ {
				num, _ := strconv.Atoi(record[j])
				line = append(line, num)
			}
			res.rank_truth = append(res.rank_truth, line)
		}
	} else {
		fmt.Println("evaluation method doesn't exist")
		os.Exit(3)
	}
	
	return res
}

func (daemon *Daemon) loadCompetitionData() {
	rows, err := daemon.postgresConn.Query(
        "SELECT c.id, c.public_truth, c.private_truth, c.num_line, c.evaluation FROM competition_competition c WHERE c.allow_overdue_submission=true or now()<c.end_datetime")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}

	daemon.competitionTruths = map[int]Competition{}
	for rows.Next() {
		var compID int
		var publicPath, privatePath string
		var numLine int
		var evaluation int
		if err := rows.Scan(&compID, &publicPath, &privatePath, &numLine, &evaluation); err != nil {
			log.Fatal(err)
		}

		daemon.competitionTruths[compID] = Competition{
			readTruth(path.Join(daemon.config.TruthRoot, publicPath), evaluation),
			readTruth(path.Join(daemon.config.TruthRoot, privatePath), evaluation),
			numLine,
			evaluation,
		}
	}

}

func (daemon *Daemon) Init(config Config) {
	daemon.config = config
	daemon.connDb()
	daemon.connRedis()
	daemon.loadCompetitionData()
}

func (daemon *Daemon) Cleanup() {
	daemon.postgresConn.Close()
	daemon.redisConn.Close()
}

func (daemon *Daemon) Run() {
	queue := make(chan Submission)
	for i := 0; i < daemon.config.Worker; i += 1 {
		go daemon.work(queue)
	}

	for {
		submissionJson, _ := redis.Bytes(daemon.redisConn.Do("LPOP", "submission_queue"))
		if submissionJson == nil {
			continue
		}
		var submission Submission
		err := json.Unmarshal(submissionJson, &submission)
		if err != nil {
			fmt.Println(err.Error())
			time.Sleep(1000 * time.Millisecond)
		} else {
			queue <- submission
		}
	}
}
func Start(configFilename string) {
	config, err := LoadConfig(configFilename)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	daemon := Daemon{}
	daemon.Init(config)
	defer daemon.Cleanup()
	daemon.Run()
}

func evaluate(truth Truth, predication Prediction) float32 {
	return NDCG(truth, predication)
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
							SET public_score=$1, private_score=$2, status=2, message=$3
							WHERE id=$4`,
		publicScore, privateScore, "Success", submissionPk)
	if err != nil {
		log.Println(err.Error())
	}
}

func (daemon *Daemon) work(queue chan Submission) {
	fmt.Println("worker")

	for {
		submission := <-queue
		evaluation_method := daemon.competitionTruths[submission.CompetitionPk].evaluation
		predication, err := submission.ReadData(evaluation_method)

		if err != nil {
			fmt.Println(err.Error())
			daemon.writeMsg(submission.Pk, err.Error())
			continue
		}

		if evaluation_method == 2 {
			if len(predication.classification_prediction) != daemon.competitionTruths[submission.CompetitionPk].NumLine {
				fmt.Println("Line number doesn't match")
				daemon.writeMsg(submission.Pk, "Line number doesn't match")
			}
		}
		if evaluation_method == 1 {
			if len(predication.rank_prediction) != daemon.competitionTruths[submission.CompetitionPk].NumLine {
				fmt.Println("Line number doesn't match")
				daemon.writeMsg(submission.Pk, "Line number doesn't match")
			}
		}	
		var publicScore, privateScore float32
		publicScore = evaluate(daemon.competitionTruths[submission.CompetitionPk].Public, predication)
		privateScore = 1.0
		fmt.Println(publicScore)
		daemon.writeScore(submission.Pk, publicScore, privateScore)
	}
}
