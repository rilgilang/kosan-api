package service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rilgilang/kosan-api/config/dotenv"
	"github.com/rilgilang/kosan-api/internal/api/presenter"
	"github.com/rilgilang/kosan-api/internal/entities"
	"github.com/rilgilang/kosan-api/internal/pkg/logger"
	"github.com/rilgilang/kosan-api/internal/repositories"
)

type RoomService interface {
	GetAllRooms(ctx context.Context) *presenter.Response
	GetDetailedRoom(ctx context.Context, roomId string) *presenter.Response
}

type roomService struct {
	roomRepo             repositories.RoomRepository
	paymentHistoriesRepo repositories.PaymentHistory
	cfg                  *dotenv.Config
}

func NewRoomService(roomRepo repositories.RoomRepository, paymentHistoriesRepo repositories.PaymentHistory, cfg *dotenv.Config) RoomService {
	return &roomService{
		roomRepo:             roomRepo,
		paymentHistoriesRepo: paymentHistoriesRepo,
		cfg:                  cfg,
	}
}

func (s *roomService) GetAllRooms(ctx context.Context) *presenter.Response {
	var (
		log      = logger.NewLog("room_handler", s.cfg.LoggerEnable)
		response = presenter.Response{}
	)

	rooms, err := s.roomRepo.FetchAll(ctx)

	if err != nil {
		log.Error(fmt.Sprintf(`error fetching room data to db got %s`, err))
		return response.WithCode(500).WithError(errors.New("something went wrong!"))
	}

	return response.WithCode(200).WithData(rooms)
}

func (s *roomService) GetDetailedRoom(ctx context.Context, roomId string) *presenter.Response {
	var (
		log      = logger.NewLog("room_handler", s.cfg.LoggerEnable)
		response = presenter.Response{}
	)

	room, err := s.roomRepo.FetchOne(ctx, roomId)

	if err != nil {
		log.Error(fmt.Sprintf(`error fetching room data to db got %s`, err))
		return response.WithCode(500).WithError(errors.New("something went wrong!"))
	}

	if room == nil {
		return response.WithCode(404).WithError(errors.New("room not found!"))
	}

	paymentsHistories, err := s.paymentHistoriesRepo.FetchAllByRoomID(ctx, roomId)
	if err != nil {
		log.Error(fmt.Sprintf(` error fetching payment histories data to db got %s`, err))
		return response.WithCode(500).WithError(errors.New("something went wrong!"))
	}

	result := entities.DetailedRoom{
		ID:                   room.ID,
		IDCard:               room.IDCard,
		RoomImage:            room.RoomImage,
		RoomNumber:           room.RoomNumber,
		Renter:               room.Renter,
		Price:                room.Price,
		AlreadyPaidThisMonth: room.AlreadyPaidThisMonth,
		Available:            room.Available,
		FirstCheckIn:         room.FirstCheckIn,
		CheckIn:              room.CheckIn,
		CheckOut:             room.CheckOut,
		PaymentHistory:       paymentsHistories,
		CreatedAt:            room.CreatedAt,
		UpdatedAt:            room.UpdatedAt,
		DeletedAt:            room.DeletedAt,
	}

	return response.WithCode(200).WithData(&result)
}

func (s *roomService) UpdateRenter(ctx context.Context, payload entities.UpdateRenterPayload) *presenter.Response {
	var (
		log      = logger.NewLog("room_handler", s.cfg.LoggerEnable)
		response = presenter.Response{}
	)

	room, err := s.roomRepo.UpdateRenter(ctx, map[string]string{
		"renter":  payload.Renter,
		"id_card": payload.IDCard,
		"id":      payload.ID,
	})

	if err != nil {
		log.Error(fmt.Sprintf(`error update room data to db got %s`, err))
		return response.WithCode(500).WithError(errors.New("something went wrong!"))
	}

	return response.WithCode(200).WithData(room)
}
