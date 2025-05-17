#!/usr/bin/env pwsh

# Test script for IPFS API documentation integration

Write-Host "Testing IPFS API Documentation Integration" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

# Step 1: Check if IPFS service is running
Write-Host "Step 1: Checking if IPFS service is running..." -ForegroundColor Green
$ipfsRunning = $null

try {
    $ipfsRunning = docker ps | Select-String "tracepost-ipfs"
} catch {
    Write-Host "Error: Docker command failed. Make sure Docker is running." -ForegroundColor Red
    exit 1
}

if (-not $ipfsRunning) {
    Write-Host "IPFS service is not running. Starting services..." -ForegroundColor Yellow
    docker-compose up -d ipfs
    Write-Host "Waiting for IPFS service to start..." -ForegroundColor Yellow
    Start-Sleep -Seconds 5
}

Write-Host "IPFS service is running." -ForegroundColor Green

# Step 2: Check if the API is running
Write-Host "Step 2: Checking if the API server is running..." -ForegroundColor Green
$apiRunning = $null

try {
    $apiRunning = docker ps | Select-String "tracepost-larvae-api"
} catch {
    Write-Host "Error checking API status" -ForegroundColor Red
}

if (-not $apiRunning) {
    Write-Host "API service is not running. Please start it manually:" -ForegroundColor Yellow
    Write-Host "docker-compose up -d api" -ForegroundColor Yellow
    $startApi = Read-Host "Do you want to start the API now? (y/n)"
    
    if ($startApi -eq "y") {
        docker-compose up -d api
        Write-Host "Waiting for API service to start..." -ForegroundColor Yellow
        Start-Sleep -Seconds 10
    } else {
        Write-Host "Exiting test script." -ForegroundColor Red
        exit 1
    }
}

# Step 3: Test the API documentation endpoint
Write-Host "Step 3: Testing the API documentation endpoint..." -ForegroundColor Green

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/docs/ipfs" -Method GET -MaximumRedirection 0 -ErrorAction SilentlyContinue
    
    if ($response.StatusCode -eq 302) {
        Write-Host "Success: API documentation endpoint is working and redirecting to IPFS WebUI" -ForegroundColor Green
        Write-Host "Redirect URL: $($response.Headers.Location)" -ForegroundColor Green
        
        Write-Host "`nTo view the API documentation on IPFS WebUI:" -ForegroundColor Cyan
        Write-Host "1. Open your browser and navigate to http://localhost:8080/docs/ipfs" -ForegroundColor Cyan
        Write-Host "2. You will be redirected to the IPFS WebUI with the API documentation" -ForegroundColor Cyan
    } else {
        Write-Host "Warning: Received unexpected response code: $($response.StatusCode)" -ForegroundColor Yellow
    }
} catch {
    if ($_.Exception.Response.StatusCode.value__ -eq 302) {
        $location = $_.Exception.Response.Headers.Location
        Write-Host "Success: API documentation endpoint is working and redirecting to IPFS WebUI" -ForegroundColor Green
        Write-Host "Redirect URL: $location" -ForegroundColor Green
        
        Write-Host "`nTo view the API documentation on IPFS WebUI:" -ForegroundColor Cyan
        Write-Host "1. Open your browser and navigate to http://localhost:8080/docs/ipfs" -ForegroundColor Cyan
        Write-Host "2. You will be redirected to the IPFS WebUI with the API documentation" -ForegroundColor Cyan
    } else {
        Write-Host "Error: Failed to connect to the API documentation endpoint. Error: $_" -ForegroundColor Red
        Write-Host "Make sure the API server is running on port 8080." -ForegroundColor Yellow
    }
}

Write-Host "`nTest completed!" -ForegroundColor Cyan
