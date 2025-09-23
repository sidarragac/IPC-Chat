# Nombre del módulo
MODULE = ipc-chat

# Directorios
SERVER_DIR = Server
CLIENT_DIR = Client

# Ejecutables
SERVER_BIN = $(SERVER_DIR)/server
CLIENT_BIN = $(CLIENT_DIR)/client

# Comandos
GO = go

all: build

# Compilar ambos binarios
build: $(SERVER_BIN) $(CLIENT_BIN)

$(SERVER_BIN): $(SERVER_DIR)/server.go
	$(GO) build -o $(SERVER_BIN) ./$(SERVER_DIR)

$(CLIENT_BIN): $(CLIENT_DIR)/client.go
	$(GO) build -o $(CLIENT_BIN) ./$(CLIENT_DIR)

# Ejecutar servidor
run-server: $(SERVER_BIN)
	./$(SERVER_BIN)

# Ejecutar cliente
run-client: $(CLIENT_BIN)
	./$(CLIENT_BIN)

# Limpiar ejecutables
clean:
	rm -f $(SERVER_BIN) $(CLIENT_BIN)
