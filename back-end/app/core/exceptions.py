from typing import Any, Dict, Optional


class BlockchainLogisticsException(Exception):
    """Base exception for all blockchain logistics application errors"""
    
    def __init__(self, detail: str):
        self.detail = detail


class ResourceNotFoundError(BlockchainLogisticsException):
    """Resource not found error"""
    pass


class ResourceAlreadyExistsError(BlockchainLogisticsException):
    """Resource already exists error"""
    pass


class AuthenticationError(BlockchainLogisticsException):
    """Authentication failed error"""
    pass


class PermissionDeniedError(BlockchainLogisticsException):
    """Permission denied error"""
    pass


class ValidationError(BlockchainLogisticsException):
    """Data validation error"""
    pass


class BlockchainError(BlockchainLogisticsException):
    """Blockchain interaction error"""
    pass


class OracleError(BlockchainLogisticsException):
    """Oracle data source error"""
    pass


class BridgeError(BlockchainLogisticsException):
    """Cross-chain bridge error"""
    pass

class CredentialsException(Exception):
    def __init__(self, detail: str = "Invalid credentials"):
        self.detail = detail

class InactiveUserException(Exception):
    def __init__(self, detail: str = "Inactive user"):
        self.detail = detail

class AlertProcessingError(BlockchainLogisticsException):
    """Alert processing error"""
    pass