package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
)

func main() {
	// Define command-line flags
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	generateEntityType := generateCmd.String("type", "", "Entity type (e.g., 'hatchery', 'farmer', 'processor')")
	generateEntityName := generateCmd.String("name", "", "Entity name")
	
	proofCmd := flag.NewFlagSet("proof", flag.ExitOnError)
	proofDID := proofCmd.String("did", "", "DID to generate proof for")
	proofKeyFile := proofCmd.String("key", "", "Path to private key file")
	
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	verifyDID := verifyCmd.String("did", "", "DID to verify")
	verifyProof := verifyCmd.String("proof", "", "Proof to verify")
	
	// Check if there are enough arguments
	if len(os.Args) < 2 {
		fmt.Println("Expected 'generate', 'proof', or 'verify' subcommands")
		os.Exit(1)
	}
	
	// Load config
	cfg := config.GetConfig()
	
	// Parse the appropriate command
	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		if *generateEntityType == "" || *generateEntityName == "" {
			fmt.Println("Entity type and name are required")
			generateCmd.PrintDefaults()
			os.Exit(1)
		}
		generateDID(cfg, *generateEntityType, *generateEntityName)
		
	case "proof":
		proofCmd.Parse(os.Args[2:])
		if *proofDID == "" || *proofKeyFile == "" {
			fmt.Println("DID and key file are required")
			proofCmd.PrintDefaults()
			os.Exit(1)
		}
		generateProof(*proofDID, *proofKeyFile)
		
	case "verify":
		verifyCmd.Parse(os.Args[2:])
		if *verifyDID == "" || *verifyProof == "" {
			fmt.Println("DID and proof are required")
			verifyCmd.PrintDefaults()
			os.Exit(1)
		}
		verifyDIDProof(cfg, *verifyDID, *verifyProof)
		
	default:
		fmt.Println("Expected 'generate', 'proof', or 'verify' subcommands")
		os.Exit(1)
	}
}

// generateDID generates a new DID and saves the key to a file
func generateDID(cfg *config.Config, entityType, entityName string) {
	fmt.Println("Generating new DID for", entityName, "of type", entityType)
	
	// Register DID
	did, privateKeyPEM, err := blockchain.RegisterDID(
		cfg.BlockchainNodeURL,
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
		entityType,
		entityName,
	)
	if err != nil {
		fmt.Println("Error generating DID:", err)
		os.Exit(1)
	}
	
	// Create a filename from the DID
	filename := strings.Replace(did, ":", "_", -1) + ".key"
	
	// Save private key to file
	err = os.WriteFile(filename, []byte(privateKeyPEM), 0600)
	if err != nil {
		fmt.Println("Error saving private key:", err)
		os.Exit(1)
	}
	
	fmt.Println("DID successfully generated:")
	fmt.Println("DID:", did)
	fmt.Println("Private key saved to:", filename)
	fmt.Println("IMPORTANT: Keep this file secure and never share it.")
}

// generateProof generates a proof for a DID using the private key
func generateProof(did, keyFile string) {
	// Read private key from file
	privateKeyPEM, err := os.ReadFile(keyFile)
	if err != nil {
		fmt.Println("Error reading private key:", err)
		os.Exit(1)
	}

	// Initialize BlockchainClient
	blockchainClient := blockchain.NewBlockchainClient(
		"http://blockchain-node-url", // Replace with actual node URL
		"", // Private key not needed for this operation
		"account-address", // Replace with actual account address
		"chain-id", // Replace with actual chain ID
		"poa", // Replace with actual consensus type
	)

	// Create DDI client
	client, err := blockchain.NewDDIClient(blockchain.DDIClientConfig{
		PrivateKeyPEM: string(privateKeyPEM),
		DID:           did,
	}, blockchainClient)
	if err != nil {
		fmt.Println("Error creating DDI client:", err)
		os.Exit(1)
	}
	
	// Generate proof
	proof, err := client.GenerateProof()
	if err != nil {
		fmt.Println("Error generating proof:", err)
		os.Exit(1)
	}
	
	// Output the proof and instructions
	fmt.Println("DID Proof successfully generated for", did)
	fmt.Println("\nProof:", proof)
	fmt.Println("\nTo use this proof for API authentication, include the following HTTP headers:")
	fmt.Println("X-DID:", did)
	fmt.Println("X-DID-Proof:", proof)
	fmt.Println("\nNOTE: This proof is only valid for a short time. Generate a new proof for each API request.")
	
	// Also output as JSON
	jsonOutput := map[string]string{
		"did":   did,
		"proof": proof,
	}
	jsonBytes, _ := json.MarshalIndent(jsonOutput, "", "  ")
	fmt.Println("\nJSON Format:")
	fmt.Println(string(jsonBytes))
}

// verifyDIDProof verifies a DID proof
func verifyDIDProof(cfg *config.Config, did, proof string) {
	fmt.Println("Verifying proof for DID:", did)
	
	// Initialize blockchain client
	blockchainClient := blockchain.NewBlockchainClient(
		cfg.BlockchainNodeURL,
		"", // Private key is not needed for verification
		cfg.BlockchainAccount,
		cfg.BlockchainChainID,
		cfg.BlockchainConsensus,
	)
	
	// Create identity client
	identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)
	
	// Verify proof
	isValid, err := identityClient.VerifyDIDProof(did, proof)
	if err != nil {
		fmt.Println("Error verifying proof:", err)
		os.Exit(1)
	}
	
	if isValid {
		fmt.Println("✓ Proof is valid")
		
		// Get permissions
		permissions, err := identityClient.GetActorPermissions(did)
		if err != nil {
			fmt.Println("Error getting permissions:", err)
		} else {
			fmt.Println("\nPermissions:")
			for permission, allowed := range permissions {
				if allowed {
					fmt.Println("✓", permission)
				} else {
					fmt.Println("✗", permission)
				}
			}
		}
	} else {
		fmt.Println("✗ Proof is invalid")
	}
}
