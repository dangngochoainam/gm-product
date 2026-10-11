USER_NAME=root
PASSWORD=Abc12345
HOST=localhost
PORT=5434
DB_NAME=product
SCHEMA_NAME=public

export DB_CONN_STRING="postgres://$USER_NAME:$PASSWORD@$HOST:$PORT/$DB_NAME?search_path=$SCHEMA_NAME&sslmode=disable"