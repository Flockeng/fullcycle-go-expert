package auction

import (
	"context"
	"fullcycle-auction_go/configuration/database/mongodb"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"
)

func TestCreateAuction_AutomaticClosure(t *testing.T) {
	ctx := context.Background()

	originalInterval := os.Getenv("AUCTION_INTERVAL")
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer func() {
		if originalInterval != "" {
			os.Setenv("AUCTION_INTERVAL", originalInterval)
		} else {
			os.Unsetenv("AUCTION_INTERVAL")
		}
	}()

	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/?authSource=admin"
	}

	mongoDB := os.Getenv("MONGODB_DB")
	if mongoDB == "" {
		mongoDB = "auctions"
	}

	os.Setenv("MONGODB_URL", mongoURL)
	os.Setenv("MONGODB_DB", mongoDB)

	database, err := mongodb.NewMongoDBConnection(ctx)
	if err != nil {
		t.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}

	repository := NewAuctionRepository(database)

	_, err = repository.Collection.DeleteMany(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Erro ao limpar a coleção: %v", err)
	}

	auction, internalErr := auction_entity.CreateAuction(
		"Produto Teste",
		"Categoria Teste",
		"Descrição do produto de teste para validação",
		auction_entity.New,
	)
	if internalErr != nil {
		t.Fatalf("Erro ao criar entidade de leilão: %v", internalErr)
	}

	if auction.Status != auction_entity.Active {
		t.Errorf("Esperado status Active, mas obteve %d", auction.Status)
	}

	internalErr = repository.CreateAuction(ctx, auction)
	if internalErr != nil {
		t.Fatalf("Erro ao criar leilão no banco: %v", internalErr)
	}

	createdAuction, internalErr := repository.FindAuctionById(ctx, auction.Id)
	if internalErr != nil {
		t.Fatalf("Erro ao buscar leilão criado: %v", internalErr)
	}

	if createdAuction.Status != auction_entity.Active {
		t.Errorf("Esperado status Active após criação, mas obteve %d", createdAuction.Status)
	}

	t.Logf("Aguardando fechamento automático do leilão (intervalo: 2s)...")
	time.Sleep(3 * time.Second)

	updatedAuction, internalErr := repository.FindAuctionById(ctx, auction.Id)
	if internalErr != nil {
		t.Fatalf("Erro ao buscar leilão após intervalo: %v", internalErr)
	}

	if updatedAuction.Status != auction_entity.Completed {
		t.Errorf("Esperado status Completed após intervalo, mas obteve %d. O fechamento automático não funcionou.", updatedAuction.Status)
	} else {
		t.Logf("Fechamento automático funcionou corretamente! Status atualizado para Completed.")
	}

	_, err = repository.Collection.DeleteMany(ctx, map[string]interface{}{})
	if err != nil {
		t.Logf("Aviso: Erro ao limpar a coleção após o teste: %v", err)
	}
}
