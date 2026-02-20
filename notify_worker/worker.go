package notify_worker

import (
	"context"
	"github.com/dontpanicw/calendar/internal/domain"
	"log"
)

//Фоновый воркер через канал:
//при создании события с напоминанием — кладём задачу в канал, воркер должен следить за временем и слать напоминания

type NotifyWorker struct {
	eventChan chan domain.Event
}

func NewNotifyWorker() *NotifyWorker {
	return &NotifyWorker{
		eventChan: make(chan domain.Event, 100),
	}
}

func (w *NotifyWorker) Start(ctx context.Context) {
	for {
		select {
		case event := <-w.eventChan:
			err := w.schedule(event)
			if err != nil {
				log.Printf("failed to schedule event: %v", err)
			}
		case <-ctx.Done():
			log.Println("notify worker stopped")
			return
		}
	}
}

func (w *NotifyWorker) SendNotify(event *domain.Event) {
	select {
	case w.eventChan <- *event:
	default:
		log.Printf("Chan is full, not sending event")
	}

}

func (w *NotifyWorker) schedule(event domain.Event) error {
	//if time.Now().After(event.Date.Add(-60 * time.Minute)) {
	//	//send notify
	//	log.Printf("The event %d will start in %v minutes.", event.EventId, event.Date)
	//	return nil
	//}
	//remindAt := event.Date.Add(-60 * time.Minute)
	//delay := time.Until(remindAt)
	//
	//// Создаем таймер на одно уведомление
	//time.AfterFunc(delay, func() {
	//	// send notify
	//	log.Printf("The event %d will start in %v minutes.", event.EventId, event.Date)
	//})

	return nil
}
