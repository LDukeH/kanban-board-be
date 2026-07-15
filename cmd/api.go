package main

type application struct {
	config config
}

type config struct {
	addr string
	dbConfig dbConfig
}

type dbConfig struct {
	dsn string
}