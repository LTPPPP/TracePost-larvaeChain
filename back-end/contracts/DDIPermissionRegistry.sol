// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";

/**
 * @title DDIPermissionRegistry
 * @dev Smart contract for managing DDI permissions and access control
 * @custom:experimental This is an experimental contract for TracePost-larvaeChain
 */
contract DDIPermissionRegistry is AccessControl, Pausable {
    using ECDSA for bytes32;

    // Role definitions
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant PERMISSION_MANAGER_ROLE =
        keccak256("PERMISSION_MANAGER_ROLE");

    // Permission structure
    struct Permission {
        string action; // The action being permitted (e.g., "create_batch", "update_status")
        string resource; // The resource being accessed (e.g., "batch", "shipment")
        uint256 expiry; // Timestamp when the permission expires (0 for no expiry)
        bool isActive; // Whether the permission is currently active
        address grantedBy; // Address that granted the permission
        uint256 grantedAt; // Timestamp when the permission was granted
    }

    // DID credential structure
    struct Credential {
        string credentialType; // Type of the credential
        string issuer; // DID of the issuer
        uint256 issuedAt; // Timestamp when it was issued
        uint256 expiresAt; // Timestamp when it expires
        bool isRevoked; // Whether the credential has been revoked
        string proofValue; // Proof value (signature)
    }

    // Mapping from DID to permissions
    mapping(string => mapping(string => mapping(string => Permission)))
        public didPermissions; // did -> action -> resource -> Permission

    // Mapping from DID to credentials
    mapping(string => Credential[]) public didCredentials;

    // List of all registered DIDs
    string[] public registeredDIDs;
    mapping(string => bool) public didExists;

    // Events
    event PermissionGranted(
        string indexed didHash,
        string did,
        string action,
        string resource,
        uint256 expiry
    );

    event PermissionRevoked(
        string indexed didHash,
        string did,
        string action,
        string resource
    );

    event CredentialIssued(
        string indexed didHash,
        string did,
        string credentialType,
        string issuer,
        uint256 expiresAt
    );

    event CredentialRevoked(
        string indexed didHash,
        string did,
        string credentialType,
        string issuer
    );

    /**
     * @dev Constructor
     */
    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
        _grantRole(PERMISSION_MANAGER_ROLE, msg.sender);
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
     * @dev Register a new DID
     * @param did The DID to register
     * @param publicKey Public key associated with the DID
     */
    function registerDID(
        string calldata did,
        string calldata publicKey
    ) external onlyRole(PERMISSION_MANAGER_ROLE) whenNotPaused {
        require(!didExists[did], "DID already registered");

        // Register the DID
        registeredDIDs.push(did);
        didExists[did] = true;

        // Hash of the DID for indexed event
        string memory didHash = _hashString(did);

        emit PermissionGranted(
            didHash,
            did,
            "register",
            "system",
            0 // No expiry
        );
    }

    /**
     * @dev Grant a permission to a DID
     * @param did The DID to grant the permission to
     * @param action The action being permitted
     * @param resource The resource being accessed
     * @param expiry Timestamp when the permission expires (0 for no expiry)
     */
    function grantPermission(
        string calldata did,
        string calldata action,
        string calldata resource,
        uint256 expiry
    ) external onlyRole(PERMISSION_MANAGER_ROLE) whenNotPaused {
        require(didExists[did], "DID not registered");

        // Create the permission
        Permission memory permission = Permission({
            action: action,
            resource: resource,
            expiry: expiry,
            isActive: true,
            grantedBy: msg.sender,
            grantedAt: block.timestamp
        });

        // Store the permission
        didPermissions[did][action][resource] = permission;

        // Hash of the DID for indexed event
        string memory didHash = _hashString(did);

        emit PermissionGranted(didHash, did, action, resource, expiry);
    }

    /**
     * @dev Revoke a permission from a DID
     * @param did The DID to revoke the permission from
     * @param action The action to revoke
     * @param resource The resource to revoke access to
     */
    function revokePermission(
        string calldata did,
        string calldata action,
        string calldata resource
    ) external onlyRole(PERMISSION_MANAGER_ROLE) whenNotPaused {
        require(didExists[did], "DID not registered");
        require(
            didPermissions[did][action][resource].isActive,
            "Permission not active"
        );

        // Revoke the permission
        didPermissions[did][action][resource].isActive = false;

        // Hash of the DID for indexed event
        string memory didHash = _hashString(did);

        emit PermissionRevoked(didHash, did, action, resource);
    }

    /**
     * @dev Check if a DID has a specific permission
     * @param did The DID to check
     * @param action The action to check
     * @param resource The resource to check
     * @return bool Whether the DID has the permission
     */
    function hasPermission(
        string calldata did,
        string calldata action,
        string calldata resource
    ) external view returns (bool) {
        if (!didExists[did]) {
            return false;
        }

        Permission memory permission = didPermissions[did][action][resource];

        // Check if permission exists, is active, and has not expired
        if (!permission.isActive) {
            return false;
        }

        if (permission.expiry != 0 && permission.expiry < block.timestamp) {
            return false;
        }

        return true;
    }

    /**
     * @dev Issue a credential to a DID
     * @param did The DID to issue the credential to
     * @param credentialType The type of credential
     * @param expiresAt Timestamp when the credential expires
     * @param proofValue Proof value (signature)
     */
    function issueCredential(
        string calldata did,
        string calldata credentialType,
        uint256 expiresAt,
        string calldata proofValue
    ) external onlyRole(PERMISSION_MANAGER_ROLE) whenNotPaused {
        require(didExists[did], "DID not registered");

        // Create the credential
        Credential memory credential = Credential({
            credentialType: credentialType,
            issuer: "did:tracepost:system", // System-issued by default
            issuedAt: block.timestamp,
            expiresAt: expiresAt,
            isRevoked: false,
            proofValue: proofValue
        });

        // Store the credential
        didCredentials[did].push(credential);

        // Hash of the DID for indexed event
        string memory didHash = _hashString(did);

        emit CredentialIssued(
            didHash,
            did,
            credentialType,
            credential.issuer,
            expiresAt
        );
    }

    /**
     * @dev Revoke a credential from a DID
     * @param did The DID to revoke the credential from
     * @param credentialIndex Index of the credential to revoke
     */
    function revokeCredential(
        string calldata did,
        uint256 credentialIndex
    ) external onlyRole(PERMISSION_MANAGER_ROLE) whenNotPaused {
        require(didExists[did], "DID not registered");
        require(
            credentialIndex < didCredentials[did].length,
            "Invalid credential index"
        );
        require(
            !didCredentials[did][credentialIndex].isRevoked,
            "Credential already revoked"
        );

        // Revoke the credential
        didCredentials[did][credentialIndex].isRevoked = true;

        // Hash of the DID for indexed event
        string memory didHash = _hashString(did);

        emit CredentialRevoked(
            didHash,
            did,
            didCredentials[did][credentialIndex].credentialType,
            didCredentials[did][credentialIndex].issuer
        );
    }

    /**
     * @dev Verify a transaction request using DID permissions
     * @param did The DID requesting the transaction
     * @param action The action being performed
     * @param resource The resource being accessed
     * @param signature Signature of the message hash
     * @param messageHash Hash of the message being signed
     * @return bool Whether the transaction is valid
     */
    function verifyTransaction(
        string calldata did,
        string calldata action,
        string calldata resource,
        bytes calldata signature,
        bytes32 messageHash
    ) external view returns (bool) {
        // Check if DID has the required permission
        if (!this.hasPermission(did, action, resource)) {
            return false;
        }

        // Verify signature (this would need implementation based on the DID method)
        // For demonstration purposes, we'll just return true if permission exists
        return true;
    }

    /**
     * @dev Get credentials for a DID
     * @param did The DID to get credentials for
     * @return Credential[] Array of credentials
     */
    function getCredentials(
        string calldata did
    ) external view returns (Credential[] memory) {
        require(didExists[did], "DID not registered");
        return didCredentials[did];
    }

    /**
     * @dev Get all registered DIDs
     * @return string[] Array of registered DIDs
     */
    function getAllDIDs() external view returns (string[] memory) {
        return registeredDIDs;
    }

    /**
     * @dev Hash a string to use in indexed events
     * @param value The string to hash
     * @return string The hashed string
     */
    function _hashString(
        string memory value
    ) internal pure returns (string memory) {
        return bytes32ToString(keccak256(abi.encodePacked(value)));
    }

    /**
     * @dev Convert bytes32 to string
     * @param _bytes32 The bytes32 to convert
     * @return string The resulting string
     */
    function bytes32ToString(
        bytes32 _bytes32
    ) internal pure returns (string memory) {
        uint8 i = 0;
        while (i < 32 && _bytes32[i] != 0) {
            i++;
        }
        bytes memory bytesArray = new bytes(i);
        for (i = 0; i < 32 && _bytes32[i] != 0; i++) {
            bytesArray[i] = _bytes32[i];
        }
        return string(bytesArray);
    }
}
