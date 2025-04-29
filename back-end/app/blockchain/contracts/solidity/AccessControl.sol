// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title AccessControl
 * @dev Controls access to supply chain operations
 */
contract AccessControl {
    // Role definitions
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant SHIPPER_ROLE = keccak256("SHIPPER_ROLE");
    bytes32 public constant CARRIER_ROLE = keccak256("CARRIER_ROLE");
    bytes32 public constant RECEIVER_ROLE = keccak256("RECEIVER_ROLE");
    bytes32 public constant VERIFIER_ROLE = keccak256("VERIFIER_ROLE");

    // Mapping from role to accounts
    mapping(bytes32 => mapping(address => bool)) private _roles;

    // Organization mapping
    mapping(address => string) private _organizations;
    mapping(string => bool) private _registeredOrganizations;

    // Events
    event RoleGranted(
        bytes32 indexed role,
        address indexed account,
        address indexed grantor
    );
    event RoleRevoked(
        bytes32 indexed role,
        address indexed account,
        address indexed revoker
    );
    event OrganizationRegistered(
        string organizationId,
        address indexed account
    );

    /**
     * @dev Contract constructor
     */
    constructor() {
        // Make the deployer an admin
        _grantRole(ADMIN_ROLE, msg.sender);
    }

    /**
     * @dev Modifier to check if an account has a specific role
     */
    modifier onlyRole(bytes32 role) {
        require(
            hasRole(role, msg.sender),
            "AccessControl: account does not have role"
        );
        _;
    }

    /**
     * @dev Check if an account has a role
     * @param role Role identifier
     * @param account Address to check
     * @return bool True if account has the role
     */
    function hasRole(bytes32 role, address account) public view returns (bool) {
        return _roles[role][account];
    }

    /**
     * @dev Grant a role to an account
     * @param role Role identifier
     * @param account Account to grant role to
     */
    function grantRole(
        bytes32 role,
        address account
    ) public onlyRole(ADMIN_ROLE) {
        _grantRole(role, account);
    }

    /**
     * @dev Internal function to grant a role
     */
    function _grantRole(bytes32 role, address account) private {
        if (!_roles[role][account]) {
            _roles[role][account] = true;
            emit RoleGranted(role, account, msg.sender);
        }
    }

    /**
     * @dev Revoke a role from an account
     * @param role Role identifier
     * @param account Account to revoke role from
     */
    function revokeRole(
        bytes32 role,
        address account
    ) public onlyRole(ADMIN_ROLE) {
        if (_roles[role][account]) {
            _roles[role][account] = false;
            emit RoleRevoked(role, account, msg.sender);
        }
    }

    /**
     * @dev Register an organization
     * @param organizationId Unique organization identifier
     * @param account Address associated with the organization
     */
    function registerOrganization(
        string memory organizationId,
        address account
    ) public onlyRole(ADMIN_ROLE) {
        require(
            !_registeredOrganizations[organizationId],
            "Organization ID already registered"
        );

        _organizations[account] = organizationId;
        _registeredOrganizations[organizationId] = true;

        emit OrganizationRegistered(organizationId, account);
    }

    /**
     * @dev Get the organization ID for an account
     * @param account Address to check
     * @return string Organization ID
     */
    function getOrganization(
        address account
    ) public view returns (string memory) {
        return _organizations[account];
    }

    /**
     * @dev Check if an organization is registered
     * @param organizationId Organization ID to check
     * @return bool True if registered
     */
    function isOrganizationRegistered(
        string memory organizationId
    ) public view returns (bool) {
        return _registeredOrganizations[organizationId];
    }
}
