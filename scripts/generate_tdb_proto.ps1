# TDB Proto Generation Script for Windows
# This script generates Go code from TDB proto files

$ErrorActionPreference = "Stop"

$ROOT_DIR = "F:\学习资料\我的论文\VDB\前置研究\code\TDB"
$PROTO_DIR = "$ROOT_DIR\pkg\proto"
$PROTOC_BIN = "protoc"

# Check if protoc is available
$protocPath = Get-Command protoc -ErrorAction SilentlyContinue
if (-not $protocPath) {
    Write-Error "protoc not found in PATH. Please install protoc first."
    exit 1
}

Write-Host "Using protoc: $($protocPath.Source)"

# Create output directories
$protoDirs = @("agentpb", "eventpb", "memorypb", "governancepb")
foreach ($dir in $protoDirs) {
    $outDir = Join-Path $PROTO_DIR $dir
    if (-not (Test-Path $outDir)) {
        New-Item -ItemType Directory -Path $outDir -Force | Out-Null
        Write-Host "Created directory: $outDir"
    }
}

# Set proto paths
$protoPath = "--proto_path=$PROTO_DIR"

# Generate agent.proto
Write-Host "Generating agent.proto..."
& $PROTOC_BIN $protoPath `
    "--go_out=paths=source_relative:$PROTO_DIR\agentpb" `
    "--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:$PROTO_DIR\agentpb" `
    "$PROTO_DIR\agent.proto"

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to generate agent.proto"
    exit 1
}

# Generate event.proto
Write-Host "Generating event.proto..."
& $PROTOC_BIN $protoPath `
    "--go_out=paths=source_relative:$PROTO_DIR\eventpb" `
    "--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:$PROTO_DIR\eventpb" `
    "$PROTO_DIR\event.proto"

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to generate event.proto"
    exit 1
}

# Generate memory.proto
Write-Host "Generating memory.proto..."
& $PROTOC_BIN $protoPath `
    "--go_out=paths=source_relative:$PROTO_DIR\memorypb" `
    "--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:$PROTO_DIR\memorypb" `
    "$PROTO_DIR\memory.proto"

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to generate memory.proto"
    exit 1
}

# Generate governance.proto
Write-Host "Generating governance.proto..."
& $PROTOC_BIN $protoPath `
    "--go_out=paths=source_relative:$PROTO_DIR\governancepb" `
    "--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:$PROTO_DIR\governancepb" `
    "$PROTO_DIR\governance.proto"

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to generate governance.proto"
    exit 1
}

Write-Host "All proto files generated successfully!"
