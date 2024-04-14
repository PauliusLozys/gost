templ:
	@templ generate

app:
	@go run ./cmd/gost

tailwind:
	@npx tailwindcss -i ./cmd/gost/assets/tailwind.css -o ./cmd/gost/assets/output.css --watch  

run: templ app

dd: # docker down
	@docker compose down

du: #d docker up
	@docker compose up -d

.PHONY: templ app tailwind dd du