// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";

/**
 * @title AccessValidator
 * @dev Smart contract for validating access before writing data to blockchain
 * @custom:experimental This is an experimental contract for TracePost-larvaeChain
 */
contract AccessValidator is AccessControl, Pausable {
    using ECDSA for bytes32;

    // Role definitions
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant VALIDATOR_ROLE = keccak256("VALIDATOR_ROLE");
    
    // Address of the DDI permission registry contract
    address public permissionRegistryAddress;
    
    // Interface for the DDI permission registry
    interface IDDIPermissionRegistry {
        function hasPermission(string calldata did, string calldata action, string calldata resource) external view returns (bool);
        function verifyTransaction(string calldata did, string calldata action, string calldata resource, bytes calldata signature, bytes32 messageHash) external view returns (bool);
    }
    
    // Mapping of transaction hashes to validation status
    mapping(bytes32 => bool) public validatedTransactions;
    
    // Events
    event TransactionValidated(bytes32 indexed txHash, string did, string action, string resource);
    event TransactionRejected(bytes32 indexed txHash, string did, string action, string resource, string reason);
    
    /**
     * @dev Constructor
     * @param _permissionRegistryAddress Address of the DDI permission registry contract
     */
    constructor(address _permissionRegistryAddress) {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
        _grantRole(VALIDATOR_ROLE, msg.sender);
        
        permissionRegistryAddress = _permissionRegistryAddress;
    }
    
    /**
     * @dev Set the permission registry address
     * @param _permissionRegistryAddress Address of the DDI permission registry contract
     */
    function setPermissionRegistryAddress(address _permissionRegistryAddress) external onlyRole(ADMIN_ROLE) {
        permissionRegistryAddress = _permissionRegistryAddress;
    }
    
    /**
     * @dev Pause the contract
     * @notice Only callable by admin
     */
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /**
     * @dev Unpause the contract
     * @notice Only callable by admin
     */
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }
    
    /**
     * @dev Validate a transaction
     * @param did The DID requesting the transaction
     * @param action The action being performed
     * @param resource The resource being accessed
     * @param signature Signature of the message hash
     * @param messageHash Hash of the message being signed
     * @return bool Whether the transaction is valid
     */
    function validateTransaction(
        string calldata did,
        string calldata action,
        string calldata resource,
        bytes calldata signature,
        bytes32 messageHash
    ) external whenNotPaused onlyRole(VALIDATOR_ROLE) returns (bool) {
        // Generate transaction hash
        bytes32 txHash = keccak256(abi.encodePacked(did, action, resource, messageHash));
        
        // Check if transaction already validated
        if (validatedTransactions[txHash]) {
            return true;
        }
        
        // Get permission registry
        IDDIPermissionRegistry permissionRegistry = IDDIPermissionRegistry(permissionRegistryAddress);
        
        // Check permission
        bool hasPermission = permissionRegistry.hasPermission(did, action, resource);
        if (!hasPermission) {
            emit TransactionRejected(txHash, did, action, resource, "Insufficient permissions");
            return false;
        }
        
        // Verify transaction
        bool isValid = permissionRegistry.verifyTransaction(did, action, resource, signature, messageHash);
        if (!isValid) {
            emit TransactionRejected(txHash, did, action, resource, "Invalid signature");
            return false;
        }
        
        // Mark transaction as validated
        validatedTransactions[txHash] = true;
        
        // Emit event
        emit TransactionValidated(txHash, did, action, resource);
        
        return true;
    }
    
    /**
     * @dev Check if a transaction has been validated
     * @param txHash Hash of the transaction
     * @return bool Whether the transaction is validated
     */
    function isTransactionValidated(bytes32 txHash) external view returns (bool) {
        return validatedTransactions[txHash];
    }
    
    /**
     * @dev Generate a transaction hash for the given parameters
     * @param did The DID requesting the transaction
     * @param action The action being performed
     * @param resource The resource being accessed
     * @param messageHash Hash of the message being signed
     * @return bytes32 The transaction hash
     */
    function generateTransactionHash(
        string calldata did,
        string calldata action,
        string calldata resource,
        bytes32 messageHash
    ) external pure returns (bytes32) {
        return keccak256(abi.encodePacked(did, action, resource, messageHash));
    }
}
