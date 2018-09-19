package buntdb

import (
	"fmt"
	"log"
	"strconv"

	"../../storage"
	"github.com/tidwall/buntdb"
)

func New(path string) *Storage {
	return &Storage{
		path: path,
	}
}

// Storage is interface structure
type Storage struct {
	path string
}

// Reset Client storage.
func (s *Storage) Reset() {
	s.SetInt(storage.TotalCountKey, 0)
	s.SetInt(storage.IosSuccessKey, 0)
	s.SetInt(storage.IosErrorKey, 0)
	s.SetInt(storage.AndroidSuccessKey, 0)
	s.SetInt(storage.AndroidErrorKey, 0)
}

func (s *Storage) GetAllKeys(keys *[]string) {
	db, _ := buntdb.Open(s.path)
	err := db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key string, value string) bool {
			*keys = append(*keys, key)
			return true
		})
		return err
	})

	if err != nil {
		log.Println("BuntDB get error:", err.Error())
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Println("BuntDB error:", err.Error())
		}
	}()
}

func (s *Storage) GetAll(values *[]string) {
	db, _ := buntdb.Open(s.path)
	err := db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key string, value string) bool {
			*values = append(*values, value)
			return true
		})
		return err
	})

	if err != nil {
		log.Println("BuntDB get error:", err.Error())
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Println("BuntDB error:", err.Error())
		}
	}()
}

// SetInt set value as int for key
func (s *Storage) SetInt(key string, count int64) {
	s.Set(key, fmt.Sprintf("%d", count))
}

// GetInt get value as int for key
func (s *Storage) GetInt(key string, count *int64) {
	var value string
	s.Get(key, &value)
	*count, _ = strconv.ParseInt(value, 10, 64)
}

// Set set value for key
func (s *Storage) Set(key string, value string) {
	db, _ := buntdb.Open(s.path)

	err := db.Update(func(tx *buntdb.Tx) error {
		if _, _, err := tx.Set(key, value, nil); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("BuntDB update error:", err.Error())
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Println("BuntDB error:", err.Error())
		}
	}()
}

// Get get value for key
func (s *Storage) Get(key string, value *string) error {
	db, _ := buntdb.Open(s.path)

	err := db.View(func(tx *buntdb.Tx) error {
		var err error
		*value, err = tx.Get(key)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println("BuntDB get error:", err.Error())
		return err
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Println("BuntDB error:", err.Error())
		}
	}()

	return nil
}
