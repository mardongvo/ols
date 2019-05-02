package main

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

//Главная структура
//Поддерживает содинение с БД, выполняет запросы
type DBKeeper struct {
	database    string //database connection string
	db          *sql.DB
	reconnectWg *sync.WaitGroup
	quit        chan bool
}

//Результат запроса
//Требется диагностировать ошибку, поэтому так обернуто
type DBResult struct {
	Error error       `json:"error"`
	Data  interface{} `json:"data"`
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func NewDBKeeper(database string) DBKeeper {
	dk := DBKeeper{database, nil, &sync.WaitGroup{}, make(chan bool)}
	go dk.checkAlive(dk.quit)
	return dk
}

func (dk *DBKeeper) Close() {
	close(dk.quit)
	if dk.db != nil {
		dk.db.Close()
	}
}

func (dk *DBKeeper) reconnect() {
	var err error
	dk.reconnectWg.Add(1)
	if dk.db != nil {
		dk.db.Close()
	}
	dk.db = nil
	defer dk.reconnectWg.Done()
	if dk.database == "" {
		log.Print("DBKeeper: строка подключения к БД пустая")
		return
	}
	for {
		dk.db, err = sql.Open("postgres", dk.database)
		if err != nil {
			log.Printf("DBKeeper: подключение к БД не удалось, ошибка %v", err)
		} else {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

//Бесконечная функция поддержки подключения к БД
func (dk *DBKeeper) checkAlive(quit <-chan bool) {
	for {
		select {
		case <-quit:
			return
		default:
			{
				if dk.db == nil {
					dk.reconnect()
				}
				if dk.db != nil {
					if err := dk.db.Ping(); err != nil {
						log.Printf("DBKeeper: ping error %v", err)
						dk.db = nil
					}
				}
				time.Sleep(5 * time.Second)
			}
		}
	}
}

//Проверка "живого" подключения
func (dk *DBKeeper) isAlive() bool {
	//ждем если вдруг идет переподключение
	//если оно слишком долгое, тогда что-то не то с БД
	if waitTimeout(dk.reconnectWg, 5*time.Second) {
		return false
	}
	return true
}
