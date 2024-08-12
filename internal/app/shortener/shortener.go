package shortener

import (
	"github.com/Kenny201/go-yandex-shortener.git/internal/domain/shortener/entity"
)

// Repository определяет интерфейс для работы с хранилищем сокращённых ссылок.
// Реализации этого интерфейса могут быть на базе различных хранилищ данных (память, файл, база данных и т.д.).
type Repository interface {
	// Get возвращает URL-объект по его идентификатору (короткому ключу).
	Get(id string) (*entity.URL, error)
	// Create создает новый короткий URL и возвращает его.
	Create(originalURL string) (string, error)
	// CreateList создает несколько коротких URL и возвращает список созданных элементов.
	CreateList(urls []*entity.URLItem) ([]*entity.URLItem, error)
	// CheckHealth проверяет состояние хранилища (доступность, целостность и т.д.).
	CheckHealth() error
}

// Shortener представляет собой основной сервис для работы с сокращёнными ссылками.
// Он использует репозиторий для сохранения и получения данных.
type Shortener struct {
	repo Repository
}

// New создает новый экземпляр сервиса Shortener с заданным репозиторием.
func New(repository Repository) *Shortener {
	return &Shortener{repo: repository}
}

// GetShortURL возвращает сокращённую ссылку по короткому ключу или ошибку, если ссылка не найдена.
func (s *Shortener) GetShortURL(shortKey string) (*entity.URL, error) {
	return s.repo.Get(shortKey)
}

// CreateShortURL сохраняет оригинальный URL в хранилище и возвращает сокращённую ссылку.
// В случае ошибки возвращает пустую ссылку и ошибку.
func (s *Shortener) CreateShortURL(originalURL string) (string, error) {
	return s.repo.Create(originalURL)
}

// CreateListShortURL сохраняет список оригинальных URL в хранилище и возвращает список сокращённых ссылок.
// В случае ошибки возвращает список частично созданных ссылок и ошибку.
func (s *Shortener) CreateListShortURL(listURL []*entity.URLItem) ([]*entity.URLItem, error) {
	return s.repo.CreateList(listURL)
}

// CheckHealth проверяет состояние репозитория, с которым работает сервис.
func (s *Shortener) CheckHealth() error {
	return s.repo.CheckHealth()
}
