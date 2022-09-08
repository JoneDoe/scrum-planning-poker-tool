package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"scrum-planning-poker-tool/logger"
	"scrum-planning-poker-tool/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/gorilla/websocket"
)

type card struct {
	ID    int `json:"id"`
	Value int `json:"value"`
}

var cardList = []card{
	card{ID: 1, Value: 1},
	card{ID: 2, Value: 2},
	card{ID: 3, Value: 3},
	card{ID: 5, Value: 5},
	card{ID: 8, Value: 8},
	card{ID: 13, Value: 13},
	card{ID: 21, Value: 21},
	card{ID: 34, Value: 34},
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Voter struct {
	ID    string
	Score string
}

type room struct {
	Number int
	Voters []*Voter
}

var rooms map[int]room

var contentDir = "src/public"

func init() {
	godotenv.Load()
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	router.LoadHTMLGlob(contentDir + "/templates/*")
	router.Static("/css", contentDir+"/css").Static("/img", contentDir+"/img").Static("/js", contentDir+"/js")

	router.GET("/", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"index.html",
			gin.H{"payload": cardList, "title": "Home Page"},
		)
	})

	router.GET("/rooms/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		//id, _ := c.Cookie("PHPSESSID")

		/*r := &rooms[123]

		r.Voters = append(r.Voters, &Voter{
			ID:    id,
			Score: score,
		})*/
		//fmt.Println(rooms)
		c.HTML(
			http.StatusOK,
			"room.html",
			gin.H{"payload": cardList, "title": "Room " + roomId},
		)
	})

	router.GET("/rooms/:roomId/results", func(c *gin.Context) {
		roomId := c.Param("roomId")

		fmt.Println(roomId)
		fmt.Println(h.rooms[roomId])
		c.HTML(
			http.StatusOK,
			"room.html",
			gin.H{"payload": cardList, "title": "Room " + roomId, "scores": h.rooms[roomId]},
		)
	})

	router.GET("/ws/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		c.SetCookie("roomsessid", "dsdfsdf", 10000, "/", "", false, false)
		fmt.Println(c.Cookie("roomsessid"))
		serveWs(c.Writer, c.Request, roomId, c)
	})

	go h.run()

	runServer(router)
}

func runServer(router *gin.Engine) {
	timeout, _ := time.ParseDuration("30s")

	server := &http.Server{
		Addr:           os.Getenv("APP_PORT"),
		Handler:        router,
		ReadTimeout:    timeout,
		WriteTimeout:   timeout,
		MaxHeaderBytes: 1 << 20,
	}

	logger.Infof("Start server on %s%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT"))

	server.ListenAndServe()
}

func serveWs2(cw http.ResponseWriter, r *http.Request, roomId string) {
	//fmt.Println(c.Cookie("PHPSESSID"))
	//Upgrade get request to webSocket protocol
	ws, err := upGrader.Upgrade(cw, r, nil)
	if err != nil {
		log.Println("error get connection")
		log.Fatal(err)
	}

	defer ws.Close()

	var data struct {
		A    string `json:"a"`
		B    int    `json:"b"`
		Room int    `json:"room"`
	}
	//Read data in ws
	err = ws.ReadJSON(&data)
	if err != nil {
		log.Println("error read json")
		log.Fatal(err)
	}

	//Write ws data, pong 10 times
	var count = 0
	for {
		count++
		if count > 5 {
			break
		}

		err = ws.WriteJSON(struct {
			A string `json:"a"`
			B int    `json:"b"`
			C int    `json:"c"`
		}{
			A: data.A,
			B: data.B,
			C: count,
		})
		if err != nil {
			log.Println("error write json: " + err.Error())
		}
		time.Sleep(1 * time.Second)
	}
}
