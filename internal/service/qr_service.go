package service

import "github.com/skip2/go-qrcode"

type QRService interface {
	GenerateQR(shortURL string) ([]byte, error)
}

type qrService struct{}

func NewQRService() QRService {
	return &qrService{}
}

func (s *qrService) GenerateQR(shortURL string) ([]byte, error) {
	return qrcode.Encode(shortURL, qrcode.Medium, 256)
}
