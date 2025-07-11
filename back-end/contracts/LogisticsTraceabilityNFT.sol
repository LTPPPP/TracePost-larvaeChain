// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Enumerable.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/utils/Counters.sol";
import "@openzeppelin/contracts/utils/Base64.sol";

/**
 * @title LogisticsTraceabilityNFT
 * @dev Smart contract for managing batch traceability using NFTs
 * @custom:experimental This is an experimental contract for TracePost-larvaeChain
 */
contract LogisticsTraceabilityNFT is
    ERC721URIStorage,
    ERC721Enumerable,
    AccessControl,
    Pausable
{
    using Counters for Counters.Counter;

    // Role definitions
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant HATCHERY_ROLE = keccak256("HATCHERY_ROLE");
    bytes32 public constant FARM_ROLE = keccak256("FARM_ROLE");
    bytes32 public constant PROCESSOR_ROLE = keccak256("PROCESSOR_ROLE");

    // Token ID counter
    Counters.Counter private _tokenIdCounter;

    // Main logistics contract address
    address public logisticsContract; // Mapping from token ID to batch ID
    mapping(uint256 => string) public tokenToBatchId;

    // Mapping from batch ID to token ID
    mapping(string => uint256) public batchIdToToken;

    // Mapping from token ID to transfer ID
    mapping(uint256 => string) public tokenToTransferId;

    // Mapping from transfer ID to token ID
    mapping(string => uint256) public transferIdToToken;

    // Mapping from token ID to transfer ID
    mapping(uint256 => string) public tokenToTransferId;

    // Mapping from transfer ID to token ID
    mapping(string => uint256) public transferIdToToken;

    // Mapping from token ID to QR code metadata
    mapping(uint256 => string) private _qrCodeMetadata;

    // Events
    event BatchTokenized(
        uint256 indexed tokenId,
        string batchId,
        string tokenURI
    );
    event BatchTransferred(
        uint256 indexed tokenId,
        string batchId,
        address from,
        address to
    );
    event BatchUpdated(
        uint256 indexed tokenId,
        string batchId,
        string newTokenURI
    );

    // Transaction NFT events
    event TransactionTokenized(
        uint256 indexed tokenId,
        string transferId,
        string batchId,
        string tokenURI
    );

    event TransactionUpdated(
        uint256 indexed tokenId,
        string transferId,
        string newTokenURI
    );

    // Events for transaction NFTs
    event TransactionTokenized(
        uint256 indexed tokenId,
        string transferId,
        string batchId,
        string tokenURI
    );

    event TransactionUpdated(
        uint256 indexed tokenId,
        string transferId,
        string newTokenURI
    );

    /**
     * @dev Constructor
     * @param _name Name of the NFT
     * @param _symbol Symbol of the NFT
     * @param _logisticsContract Address of the main logistics contract
     */
    constructor(
        string memory _name,
        string memory _symbol,
        address _logisticsContract
    ) ERC721(_name, _symbol) {
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _setupRole(ADMIN_ROLE, msg.sender);
        _setupRole(MINTER_ROLE, msg.sender);

        logisticsContract = _logisticsContract;
    }

    /**
     * @dev Create a new NFT for a batch
     * @param batchId Unique identifier of the batch
     * @param recipient Address to receive the NFT
     * @param metadataURI URI for the batch metadata (IPFS or HTTP)
     */
    function mintBatchNFT(
        string memory batchId,
        address recipient,
        string memory metadataURI
    ) public whenNotPaused onlyRole(MINTER_ROLE) returns (uint256) {
        require(batchIdToToken[batchId] == 0, "Batch already tokenized");

        // Get the next token ID
        _tokenIdCounter.increment();
        uint256 tokenId = _tokenIdCounter.current();

        // Mint the token
        _safeMint(recipient, tokenId);
        _setTokenURI(tokenId, metadataURI);

        // Update mappings
        tokenToBatchId[tokenId] = batchId;
        batchIdToToken[batchId] = tokenId;

        emit BatchTokenized(tokenId, batchId, metadataURI);

        return tokenId;
    }

    /**
     * @dev Create a new NFT for a transaction
     * @param transferId Unique identifier of the transfer/transaction
     * @param batchId Identifier of the related batch
     * @param recipient Address to receive the NFT
     * @param metadataURI URI for the transaction metadata (IPFS or HTTP)
     * @param tokenIdSuffix Optional suffix to create deterministic token IDs
     */
    function mintTransactionNFT(
        string memory transferId,
        string memory batchId,
        address recipient,
        string memory metadataURI,
        string memory tokenIdSuffix
    ) public whenNotPaused onlyRole(MINTER_ROLE) returns (uint256) {
        require(
            transferIdToToken[transferId] == 0,
            "Transaction already tokenized"
        );

        // Get the next token ID or create a deterministic one based on suffix
        uint256 tokenId;
        if (bytes(tokenIdSuffix).length > 0) {
            // Create a deterministic token ID by hashing the transfer ID and suffix
            tokenId =
                uint256(
                    keccak256(abi.encodePacked(transferId, tokenIdSuffix))
                ) %
                1000000000;
            // Ensure the token ID doesn't exist yet
            require(_exists(tokenId) == false, "Token ID already exists");
        } else {
            _tokenIdCounter.increment();
            tokenId = _tokenIdCounter.current();
        }

        // Mint the token
        _safeMint(recipient, tokenId);
        _setTokenURI(tokenId, metadataURI);

        // Update mappings
        tokenToTransferId[tokenId] = transferId;
        transferIdToToken[transferId] = tokenId;

        emit TransactionTokenized(tokenId, transferId, batchId, metadataURI);
        return tokenId;
    }

    /**
     * @dev Create a new NFT for a transaction
     * @param transferId Unique identifier of the transfer/transaction
     * @param batchId Identifier of the related batch
     * @param recipient Address to receive the NFT
     * @param metadataURI URI for the transaction metadata (IPFS or HTTP)
     * @param tokenIdSuffix Optional suffix to create deterministic token IDs
     */
    function mintTransactionNFT(
        string memory transferId,
        string memory batchId,
        address recipient,
        string memory metadataURI,
        string memory tokenIdSuffix
    ) public whenNotPaused onlyRole(MINTER_ROLE) returns (uint256) {
        require(
            transferIdToToken[transferId] == 0,
            "Transaction already tokenized"
        );

        // Get the next token ID or create a deterministic one based on suffix
        uint256 tokenId;
        if (bytes(tokenIdSuffix).length > 0) {
            // Create a deterministic token ID by hashing the transfer ID and suffix
            tokenId =
                uint256(
                    keccak256(abi.encodePacked(transferId, tokenIdSuffix))
                ) %
                1000000000;
            // Ensure the token ID doesn't exist yet
            require(_exists(tokenId) == false, "Token ID already exists");
        } else {
            _tokenIdCounter.increment();
            tokenId = _tokenIdCounter.current();
        }

        // Mint the token
        _safeMint(recipient, tokenId);
        _setTokenURI(tokenId, metadataURI);

        // Update mappings
        tokenToTransferId[tokenId] = transferId;
        transferIdToToken[transferId] = tokenId;

        emit TransactionTokenized(tokenId, transferId, batchId, metadataURI);

        return tokenId;
    }

    /**
     * @dev Generate on-chain metadata for a batch
     * @param batchId Batch identifier
     * @param species Species name
     * @param origin Origin location
     * @param timestamp Creation timestamp
     * @param imageURI URI to the batch image (QR code)
     */
    function generateTokenURI(
        string memory batchId,
        string memory species,
        string memory origin,
        uint256 timestamp,
        string memory imageURI
    ) public pure returns (string memory) {
        bytes memory metadata = abi.encodePacked(
            "{",
            '"name": "Batch #',
            batchId,
            '",',
            '"description": "Shrimp larvae batch with full traceability",',
            '"species": "',
            species,
            '",',
            '"origin": "',
            origin,
            '",',
            '"timestamp": ',
            timestamp.toString(),
            ",",
            '"image": "',
            imageURI,
            '"',
            "}"
        );

        return
            string(
                abi.encodePacked(
                    "data:application/json;base64,",
                    Base64.encode(metadata)
                )
            );
    }

    /**
     * @dev Update the metadata URI for a batch NFT
     * @param batchId Batch identifier
     * @param newMetadataURI New metadata URI
     */
    function updateBatchMetadata(
        string memory batchId,
        string memory newMetadataURI
    ) public whenNotPaused {
        uint256 tokenId = batchIdToToken[batchId];
        require(tokenId != 0, "Batch not tokenized");
        require(
            _isApprovedOrOwner(msg.sender, tokenId) ||
                hasRole(ADMIN_ROLE, msg.sender),
            "Not authorized"
        );

        _setTokenURI(tokenId, newMetadataURI);

        emit BatchUpdated(tokenId, batchId, newMetadataURI);
    }

    /**
     * @dev Transfer a batch NFT to a new owner
     * @param to Recipient address
     * @param batchId Batch identifier
     */
    function transferBatch(
        address to,
        string memory batchId
    ) public whenNotPaused {
        uint256 tokenId = batchIdToToken[batchId];
        require(tokenId != 0, "Batch not tokenized");
        require(
            _isApprovedOrOwner(msg.sender, tokenId),
            "Not owner or approved"
        );

        address from = ownerOf(tokenId);
        _safeTransfer(from, to, tokenId, "");

        emit BatchTransferred(tokenId, batchId, from, to);
    }

    /**
     * @dev Get batch ID from token ID
     * @param tokenId Token identifier
     */
    function getBatchId(uint256 tokenId) public view returns (string memory) {
        require(_exists(tokenId), "Token does not exist");
        return tokenToBatchId[tokenId];
    }

    /**
     * @dev Get token ID from batch ID
     * @param batchId Batch identifier
     */
    function getTokenId(string memory batchId) public view returns (uint256) {
        uint256 tokenId = batchIdToToken[batchId];
        require(tokenId != 0, "Batch not tokenized");
        return tokenId;
    }

    /**
     * @dev Get token information for verification
     * @param tokenId The ID of the token to query
     */
    function getTokenInfo(
        uint256 tokenId
    )
        public
        view
        returns (
            address owner,
            string memory uri,
            string memory batchId,
            string memory transferId,
            bool exists
        )
    {
        exists = _exists(tokenId);
        if (!exists) {
            return (address(0), "", "", "", false);
        }

        owner = ownerOf(tokenId);
        uri = tokenURI(tokenId);
        batchId = tokenToBatchId[tokenId];
        transferId = tokenToTransferId[tokenId];

        return (owner, uri, batchId, transferId, true);
    }

    /**
     * @dev Set the main logistics contract address
     * @param _logisticsContract New logistics contract address
     */
    function setLogisticsContract(
        address _logisticsContract
    ) public onlyRole(ADMIN_ROLE) {
        logisticsContract = _logisticsContract;
    }

    /**
     * @dev Pause the contract
     */
    function pause() public onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /**
     * @dev Unpause the contract
     */
    function unpause() public onlyRole(ADMIN_ROLE) {
        _unpause();
    }

    // Add QR code metadata to NFT contract
    function setQRCodeMetadata(uint256 tokenId, string memory qrCode) public {
        require(_exists(tokenId), "NFT does not exist");
        require(ownerOf(tokenId) == msg.sender, "Only owner can set QR code");
        _qrCodeMetadata[tokenId] = qrCode;
    }

    function getQRCodeMetadata(
        uint256 tokenId
    ) public view returns (string memory) {
        require(_exists(tokenId), "NFT does not exist");
        return _qrCodeMetadata[tokenId];
    }

    // Override functions required by inherited contracts
    function _beforeTokenTransfer(
        address from,
        address to,
        uint256 tokenId,
        uint256 batchSize
    ) internal override(ERC721, ERC721Enumerable) whenNotPaused {
        super._beforeTokenTransfer(from, to, tokenId, batchSize);
    }
    function _burn(
        uint256 tokenId
    ) internal override(ERC721, ERC721URIStorage) {
        super._burn(tokenId);

        // Clear mappings when token is burned
        string memory batchId = tokenToBatchId[tokenId];
        string memory transferId = tokenToTransferId[tokenId];

        if (bytes(batchId).length > 0) {
            delete batchIdToToken[batchId];
            delete tokenToBatchId[tokenId];
        }

        if (bytes(transferId).length > 0) {
            delete transferIdToToken[transferId];
            delete tokenToTransferId[tokenId];
        }
    }

    function tokenURI(
        uint256 tokenId
    ) public view override(ERC721, ERC721URIStorage) returns (string memory) {
        return super.tokenURI(tokenId);
    }

    function supportsInterface(
        bytes4 interfaceId
    )
        public
        view
        override(ERC721, ERC721Enumerable, AccessControl)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }
}
