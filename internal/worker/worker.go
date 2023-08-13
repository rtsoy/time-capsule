package worker

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"time-capsule/config"
	"time-capsule/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
)

func Run(ctx context.Context, cfg *config.Config, repository *repository.Repository) error {
	log.Println("(worker) started")

	for {
		time.Sleep(time.Second * 5) // Todo: Minute / Hour / Day ?

		expiredCapsules, err := repository.GetCapsules(ctx, bson.M{
			"openAt": bson.M{
				"$lte": time.Now().UTC(),
			},
		})
		if err != nil {
			log.Fatalf("(worker) failed to retrieve capsules: %s", err)
		}

		if len(expiredCapsules) == 0 {
			continue
		}

		for _, capsule := range expiredCapsules {
			user, err := repository.GetUser(ctx, bson.M{
				"_id": capsule.UserID,
			})
			if err != nil {
				fmt.Printf("(worker) failed to find user with id=%s: %s\n", user.ID.Hex(), err) // ? fatal
				continue
			}

			if err := sendEmail(cfg, "TODO???", "?", []string{user.Email}); err != nil {
				log.Println(err) // ? fatal
				continue
			}

			if err := repository.DeleteCapsule(ctx, capsule.ID); err != nil {
				log.Printf("(worker) failed to delete capsule from db with id=%s: %s\n", capsule.ID.Hex(), err) // ? fatal
				continue
			}
		}
	}

	return nil
}

func sendEmail(cfg *config.Config, subject, body string, to []string) error {
	auth := smtp.PlainAuth(
		"",
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPHost,
	)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	if err := smtp.SendMail(
		fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.SMTPUsername,
		to,
		[]byte(subject+mime+body),
	); err != nil {
		return fmt.Errorf("(worker) failed to send an email: %s", err)
	}

	return nil
}
