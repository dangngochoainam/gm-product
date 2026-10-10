USER_NAME=root
PASSWORD=Abc12345
HOST=postgres
PORT=5432
DB_NAME=pharmacy_db
SCHEMA_NAME=product

export DB_CONN_STRING="postgres://$USER_NAME:$PASSWORD@$HOST:$PORT/$DB_NAME?search_path=$SCHEMA_NAME&sslmode=disable"