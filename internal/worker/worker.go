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

const workerInterval = 5 * time.Second

// Run periodically checks for expired time capsules, retrieves the associated user information,
// and sends an email notification to users when their capsules are opened.
func Run(ctx context.Context, cfg *config.Config, repository *repository.Repository) {
	for {
		time.Sleep(workerInterval) // Todo: Minute / Hour / Day ?

		expiredCapsules, err := repository.GetCapsules(ctx, bson.M{
			"openAt": bson.M{
				"$lte": time.Now().UTC(),
			},
			"notified": false,
		})
		if err != nil {
			log.Printf("(worker) failed to retrieve capsules: %s\n", err)
		}

		fmt.Println("expired", expiredCapsules)

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

			// ? Change in the same style as the front-end design ?
			body := fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif; background-color: #f7f7f7; margin: 0; padding: 0;">
			<table align="center" border="0" cellpadding="0" cellspacing="0" width="100%s" style="max-width: 600px; margin: 20px auto; border-collapse: collapse; box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.1);">
				<tr>
					<td style="background-color: #000; padding: 40px 20px; text-align: center;">
						<h1 style="color: #ffffff; font-size: 28px;">💌 Your Time Capsule Has Been Opened</h1>
					</td>
				</tr>
				<tr>
					<td style="background-color: #ffffff; padding: 40px 40px;">
						<p style="color: #333333; font-size: 18px; line-height: 1.5;">Dear %s,</p>
						<p style="color: #333333; font-size: 18px; line-height: 1.5;">We're thrilled to share that the moment you've been waiting for has arrived. Your time capsule has been opened, revealing the cherished memories and heartfelt messages you've kept safe.</p>
						<p style="color: #333333; font-size: 18px; line-height: 1.5;">Take your time to immerse yourself in the past and relive those beautiful moments. The past is a treasure trove of emotions, and we're honored to be a part of this journey with you.</p>
						<p style="color: #333333; font-size: 18px; line-height: 1.5;">Thank you for sharing these memories with us. Here's to celebrating the richness of life and the stories that shape us.</p>
					</td>
				</tr>
				</tr>
			</table>
			</body>
			</html>
            `, "%", user.Username)

			if err = sendEmail(cfg, "Time Capsule Opened!", body, []string{user.Email}); err != nil {
				log.Println(err)
				continue
			}

			if err = repository.UpdateCapsule(ctx, capsule.ID, bson.M{
				"$set": bson.M{
					"notified": true,
				},
			}); err != nil {
				log.Println(err)
				continue
			}

			fmt.Println("email sent")
		}
	}
}

func sendEmail(cfg *config.Config, subject, body string, to []string) error {
	auth := smtp.PlainAuth(
		"",
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.SMTPHost,
	)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	subject = "Subject:" + subject + "\n"

	if err := smtp.SendMail(
		fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort),
		auth,
		cfg.SMTPUsername,
		to,
		[]byte(subject+mime+body),
	); err != nil {
		return fmt.Errorf("(worker-email) failed to send an email: %s\n", err)
	}

	return nil
}
