all: build

build:
	@docker compose -f ./srcs/config/docker-compose.yml --env-file srcs/config/.env up -d --build

down:
	@docker compose -f ./srcs/config/docker-compose.yml --env-file srcs/config/.env down

re: down clean build

docs:
	cd ./srcs/requirements/app && swag init --dir ./cmd,./internal/handler --parseDependency --parseInternal --output ./docs
clean: down
	@docker system prune -a --force

fclean:
	@docker stop $$(docker ps -qa)
	@docker system prune --all --force --volumes


.PHONY	: all build down re clean fclean

# con_to_db:
# 	psql -h localhost -U user_db -d subscription_db

#docker exec -it db_con psql -U user_db -d subscriptions_db