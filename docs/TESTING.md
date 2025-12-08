# KubeBridge-MCP Testing Guide

This document provides step-by-step instructions for testing KubeBridge-MCP with a Kubernetes cluster.

## Prerequisites

Before testing, ensure you have:

1. **Go 1.24+** installed
2. **A Kubernetes cluster** accessible via kubeconfig
3. **kubectl** configured and working with your cluster
4. **ChatGPT with MCP support** - Required for full testing

## Step 1: Build the Binary

### Option A: Using Make (Recommended)

```bash
# Navigate to the project directory
cd /path/to/KubeBridge-MCP

# Build the binary (creates ./bin/kubebridge-mcp with executable permissions)
make build

# Verify the binary was created
ls -lh bin/kubebridge-mcp
```

The `make build` command will:
- Create the `./bin` directory (if it doesn't exist)
- Build the binary to `./bin/kubebridge-mcp`
- Set executable permissions on the binary (`chmod +x`)
- The `bin/` directory is gitignored and won't be committed

### Option B: Using Go directly

```bash
# Navigate to the project directory
cd /path/to/KubeBridge-MCP

# Create bin directory
mkdir -p bin

# Build the binary
go build -o bin/kubebridge-mcp ./cmd/kubebridge-mcp

# Set executable permissions
chmod +x bin/kubebridge-mcp

# Verify the binary was created
ls -lh bin/kubebridge-mcp
```

## Step 2: Verify Kubernetes Access

Before running KubeBridge-MCP, verify that your kubeconfig is working:

```bash
# Test kubectl access
kubectl cluster-info

# List some resources to verify permissions
kubectl get namespaces
kubectl get pods -A
```

If these commands work, KubeBridge-MCP should be able to connect to your cluster.

## Step 3: Basic Connectivity Test

### Test 1: Start the Server (Read-Only Mode)

```bash
# Start KubeBridge-MCP in read-only mode
./bin/kubebridge-mcp --cluster-name test-cluster
```

**Expected Output:**
- You should see connection messages on stderr:
  - "Connected to Kubernetes cluster: test-cluster"
  - "Starting KubeBridge-MCP server..."
  - "Write operations are disabled (use --enable-write-operations to enable)"
- The server will then **wait for JSON-RPC messages on stdin**

**Important Notes:**
- ✅ If you see the connection messages, the server is working correctly!
- ⚠️ The server is **not meant to run standalone** - it needs an MCP client
- ⚠️ After the startup messages, the server will appear to "hang" - this is normal!
- ⚠️ The server is waiting for JSON-RPC protocol messages from an MCP client
- ⚠️ You may see an error like `invalid character 'D' looking for beginning of value` if you type anything - this is expected because the server expects JSON-RPC, not terminal input

**What This Test Verifies:**
- ✅ Server binary is executable and runs
- ✅ Server can connect to Kubernetes cluster
- ✅ Server discovers resources successfully
- ✅ Server is ready to accept MCP client connections

**To stop the server:** Press `Ctrl+C`

**Next Steps:**
- The server is working correctly! Now you need to configure an MCP client to use it (see Test 2)

**If you see connection errors:**
- Check your kubeconfig: `kubectl config view`
- Verify cluster access: `kubectl get nodes`
- Check RBAC permissions if using a service account

### Test 2: Understanding MCP Server Behavior

**Important**: KubeBridge-MCP is an MCP (Model Context Protocol) server that communicates via JSON-RPC over stdin/stdout. It is **not designed to run standalone** - it must be launched by an MCP client (like ChatGPT with MCP support).

When you run the server directly from the command line:
```bash
./bin/kubebridge-mcp --cluster-name test-cluster
```

**What happens:**
1. The server connects to your Kubernetes cluster ✅
2. The server discovers available resources ✅
3. The server generates MCP tools for discovered resources ✅
4. The server then **waits for JSON-RPC messages on stdin** ⏳

**The error you see** (`invalid character 'D' looking for beginning of value`) occurs because:
- The server is waiting for JSON-RPC protocol messages on stdin
- When run directly, stdin is connected to your terminal, not an MCP client
- The server expects JSON-RPC format, not terminal input

**This is normal behavior!** The server is working correctly - it just needs to be run by an MCP client.

**To verify the server works correctly**, you need to:
1. Install and configure ChatGPT with MCP support (see Step 4 below)
2. ChatGPT will launch the server and communicate with it via JSON-RPC
3. You can then test the tools through ChatGPT's interface

## Step 4: Install and Configure ChatGPT with MCP

MCP (Model Context Protocol) is a protocol that allows AI assistants to interact with external tools. To test KubeBridge-MCP, you need ChatGPT with MCP support enabled.

### Prerequisites

- ChatGPT account with MCP support (check if your ChatGPT installation supports MCP)
- ChatGPT Desktop app or compatible ChatGPT client with MCP support

### Installation and Configuration Steps:

1. **Verify ChatGPT MCP Support**
   
   - Ensure you have a ChatGPT client that supports MCP
   - MCP support may be available in:
     - ChatGPT Desktop app (if MCP is enabled)
     - Custom ChatGPT clients with MCP integration
     - Check ChatGPT documentation for MCP support status

2. **Find Your MCP Configuration File**
   
   The location varies depending on your ChatGPT installation:
   
   **macOS (ChatGPT Desktop):**
   ```bash
   ~/Library/Application Support/ChatGPT/mcp_settings.json
   # OR
   ~/Library/Application Support/OpenAI/ChatGPT/mcp_settings.json
   ```
   
   **Windows (ChatGPT Desktop):**
   ```
   %APPDATA%\ChatGPT\mcp_settings.json
   # OR
   %APPDATA%\OpenAI\ChatGPT\mcp_settings.json
   ```
   
   **Linux:**
   ```bash
   ~/.config/ChatGPT/mcp_settings.json
   # OR
   ~/.config/OpenAI/ChatGPT/mcp_settings.json
   ```
   
   **Alternative locations to check:**
   - If using VS Code with ChatGPT extension: Check extension settings
   - If using a custom client: Check that client's documentation
   - Look for files named: `mcp_settings.json`, `mcp_config.json`, `cline_mcp_settings.json`

3. **Get the Absolute Path to Your Binary**
   
   ```bash
   # From the KubeBridge-MCP directory
   pwd
   # Example output: /Users/yourname/go/src/github.com/gopherstein/KubeBridge-MCP
   # So the binary path would be: /Users/yourname/go/src/github.com/gopherstein/KubeBridge-MCP/bin/kubebridge-mcp
   
   # Verify the binary exists
   ls -lh bin/kubebridge-mcp
   # Should show: -rwxr-xr-x (executable permissions)
   ```

4. **Create or Edit Configuration File**
   
   If the file doesn't exist, create it. If it exists, edit it to add the KubeBridge-MCP server.
   
   **Add KubeBridge-MCP Configuration:**
   ```json
   {
     "mcpServers": {
       "kubebridge": {
         "command": "/absolute/path/to/KubeBridge-MCP/bin/kubebridge-mcp",
         "args": ["--cluster-name", "test-cluster"],
         "env": {
           "KUBEBRIDGE_USE_ROUTER": "true"
         }
       }
     }
   }
   ```
   
   **Note**: Router mode (`KUBEBRIDGE_USE_ROUTER: "true"`) is recommended to work around MCP client tool limits (e.g., Cursor's 40-tool limit). When enabled, only 2 router tools are exposed: `k8s_retrieve_tools` and `k8s_call_tool`.
   
   **Important**: 
   - Replace `/absolute/path/to/KubeBridge-MCP` with your actual absolute path
   - Use forward slashes (`/`) even on Windows
   - The path must be absolute, not relative
   - Ensure the JSON is valid (use a JSON validator if needed)

5. **Restart ChatGPT**
   - Quit ChatGPT completely
   - Relaunch ChatGPT
   - The MCP server should now be connected

6. **Verify Connection**
   - In ChatGPT/Cursor, start a new conversation
   - **If using router mode** (recommended):
     - Ask: "What Kubernetes tools are available?"
     - The AI should use `k8s_retrieve_tools` to discover tools
     - Ask: "List pods in the default namespace"
     - The AI should use `k8s_call_tool` with `k8s_list_pods`
   - **If using direct mode**:
     - Ask: "What Kubernetes tools are available?"
     - ChatGPT should list tools like `k8s_list_pods`, `k8s_get_pod`, etc.
   - If the tools work, the connection is working! ✅

### Troubleshooting MCP Client Connection

**If ChatGPT can't connect to the MCP server:**

1. **Check Binary Path**
   ```bash
   # Verify the binary exists and is executable
   ls -lh /absolute/path/to/KubeBridge-MCP/bin/kubebridge-mcp
   # Should show: -rwxr-xr-x (executable permissions)
   ```

2. **Test Binary Manually**
   ```bash
   # Run with --help to verify it works
   /absolute/path/to/KubeBridge-MCP/bin/kubebridge-mcp --help
   ```

3. **Check MCP Client Logs**
   - ChatGPT Desktop: Check console/logs for error messages
   - Look for JSON parsing errors or connection issues
   - Verify MCP is enabled in ChatGPT settings

4. **Verify JSON Syntax**
   - Ensure your MCP config file has valid JSON
   - Use a JSON validator if needed

5. **Check Permissions**
   ```bash
   # Ensure binary is executable
   chmod +x /absolute/path/to/KubeBridge-MCP/bin/kubebridge-mcp
   ```

## Step 5: Test with MCP Client

Now that you have ChatGPT configured and connected to KubeBridge-MCP, you can test the tools!

### Test 3: Verify Tool Discovery

In ChatGPT/Cursor:

**If using router mode** (recommended):
- Ask: "What Kubernetes tools are available?"
- The AI should use `k8s_retrieve_tools` to search for tools
- Ask: "Search for pod-related tools"
- The AI should find tools like `k8s_list_pods`, `k8s_get_pod`, etc.

**If using direct mode**:
- Ask: "What Kubernetes tools are available?"
- The AI should list tools like:
  - `k8s_list_pods`
  - `k8s_get_pod`
  - `k8s_describe_pod`
  - `k8s_list_deployments`
  - `k8s_get_deployment`
  - `k8s_get_pod_logs`
  - `k8s_get_events`
  - And many more (one for each Kubernetes resource type)

**If you see the tools listed or can use them**, the MCP connection is working correctly! ✅

## Step 6: Test Read-Only Operations

### Test 4: List Resources

In ChatGPT, test listing resources:

**Example prompts:**
- "List all pods in the default namespace"
- "Show me all deployments"
- "List all namespaces"

**Expected Results:**
- **Router mode**: AI uses `k8s_call_tool` with `k8s_list_pods` or `k8s_list_deployments`
- **Direct mode**: Tools like `k8s_list_pods`, `k8s_list_deployments` are called directly
- Resources should be returned as JSON
- No errors should occur

### Test 5: Get Resource Details

**Example prompts:**
- "Get details of pod X in namespace Y"
- "Show me the deployment named Z"
- "Describe the service called my-service"

**Expected Results:**
- **Router mode**: AI uses `k8s_call_tool` with tool names like `k8s_get_pod`, `k8s_get_deployment`
- **Direct mode**: Tools like `k8s_get_pod`, `k8s_get_deployment` are called directly
- Resource manifests should be returned
- `k8s_describe_*` tools should return formatted descriptions

### Test 6: Get Pod Logs

**Example prompts:**
- "Get logs from pod X in namespace Y"
- "Show me the last 50 lines of logs from pod Z"
- "Get logs from container 'app' in pod X"

**Expected Results:**
- **Router mode**: AI uses `k8s_call_tool` with `k8s_get_pod_logs`
- **Direct mode**: `k8s_get_pod_logs` tool is called directly
- Logs should be returned as an array of strings
- Error analysis summary should be included in text output
- Tail lines limit should be respected (max 1000)

### Test 7: Get Events

**Example prompts:**
- "Show me events in the default namespace"
- "Get events for pod X"
- "List events for deployments"

**Expected Results:**
- `k8s_get_events` tool should work
- Events should be returned in an object with an `events` field
- Events should be sorted by time (most recent first)
- Filtering by kind/name should work

## Step 7: Test Write Operations (Optional - Use with Caution!)

⚠️ **WARNING**: Enabling write operations allows ChatGPT to modify your Kubernetes cluster. Only test this in a non-production environment!

### Test 8: Enable Write Operations

```bash
# Stop the server (Ctrl+C) and restart with write operations enabled
./bin/kubebridge-mcp --enable-write-operations --cluster-name test-cluster
```

**Expected Output:**
- You should see: "WARNING: Write operations are ENABLED"

### Test 9: Create Resource

**Example prompts:**
- "Create a ConfigMap named test-cm in the default namespace with data key=value"
- "Create a simple pod manifest"

**Expected Results:**
- `k8s_create_*` tools should be available
- Resources should be created successfully
- Created resources should be visible via `kubectl`

### Test 10: Update Resource

**Example prompts:**
- "Update the ConfigMap test-cm to add a new key"
- "Modify the deployment X to change the image"

**Expected Results:**
- `k8s_update_*` tools should work
- Resources should be updated successfully
- Changes should be visible via `kubectl`

### Test 11: Patch Resource

**Example prompts:**
- "Patch the deployment X to change the replica count to 3"
- "Update the labels on pod Y"

**Expected Results:**
- `k8s_patch_*` tools should work
- Resources should be patched successfully

### Test 12: Delete Resource

**Example prompts:**
- "Delete the ConfigMap test-cm"
- "Remove the pod X"

**Expected Results:**
- `k8s_delete_*` tools should work
- Resources should be deleted successfully
- Resources should no longer exist via `kubectl`

### Test 13: Apply Resource

**Example prompts:**
- "Apply this manifest to create/update a resource"
- "Use server-side apply to update the deployment"

**Expected Results:**
- `k8s_apply_*` tools should work
- Resources should be applied successfully
- Server-side apply should handle conflicts correctly

## Step 8: Test Error Handling

### Test 14: Test RBAC Errors

If you have limited RBAC permissions, test error handling:

**Expected Behavior:**
- Operations on resources you don't have access to should fail gracefully
- Error messages should be user-friendly
- No sensitive information should be exposed

### Test 15: Test Invalid Inputs

**Example prompts:**
- "Get pod with empty name"
- "List pods in non-existent namespace"
- "Create resource with invalid manifest"

**Expected Behavior:**
- Validation errors should be clear
- Server should not crash
- Error messages should be helpful

## Step 9: Test Custom Resources (CRDs)

If your cluster has Custom Resource Definitions:

### Test 16: Verify CRD Tools

**Expected Behavior:**
- Tools should be automatically generated for CRDs
- `k8s_list_{crd}`, `k8s_get_{crd}` should work
- CRD resources should be discoverable and accessible

**Example:**
If you have a CRD called `MyCustomResource`, you should see:
- `k8s_list_mycustomresources`
- `k8s_get_mycustomresource`
- etc.

## Troubleshooting

### Issue: Server fails to start

**Check:**
1. Kubeconfig path is correct: `kubectl config view`
2. Cluster is accessible: `kubectl cluster-info`
3. Permissions are sufficient: `kubectl auth can-i list pods --all-namespaces`

### Issue: No tools are generated

**Check:**
1. Discovery API is accessible: `kubectl api-resources`
2. RBAC allows discovery: `kubectl auth can-i list --all-namespaces`
3. Check server logs for discovery errors

### Issue: Tools don't work

**Check:**
1. Resource exists: `kubectl get {resource} {name} -n {namespace}`
2. RBAC permissions: `kubectl auth can-i get {resource} -n {namespace}`
3. Resource name is correct (case-sensitive)

### Issue: Write operations fail

**Check:**
1. Write operations are enabled: Look for "WARNING: Write operations are ENABLED"
2. RBAC allows write: `kubectl auth can-i create {resource} -n {namespace}`
3. Resource manifest is valid

### Issue: ChatGPT can't connect to MCP server

**Check:**
1. Binary path is absolute in MCP config
2. Binary is executable: `chmod +x /path/to/bin/kubebridge-mcp`
3. ChatGPT MCP support is enabled
4. ChatGPT logs for connection errors
5. MCP configuration file has valid JSON syntax
6. Server binary works: `/path/to/bin/kubebridge-mcp --help`

## Test Checklist

Use this checklist to verify all functionality:

### Basic Functionality
- [ ] Server starts successfully
- [ ] Connects to Kubernetes cluster
- [ ] Discovers resources without errors
- [ ] Generates tools for discovered resources

### Read Operations
- [ ] `k8s_list_*` tools work
- [ ] `k8s_get_*` tools work
- [ ] `k8s_describe_*` tools work
- [ ] `k8s_get_pod_logs` works
- [ ] `k8s_get_events` works

### Write Operations (if enabled)
- [ ] `k8s_create_*` tools work
- [ ] `k8s_update_*` tools work
- [ ] `k8s_patch_*` tools work
- [ ] `k8s_delete_*` tools work
- [ ] `k8s_apply_*` tools work

### Error Handling
- [ ] RBAC errors are handled gracefully
- [ ] Invalid inputs return clear errors
- [ ] No sensitive information in error messages

### Advanced
- [ ] Custom Resource Definitions are discovered
- [ ] CRD tools work correctly
- [ ] Multiple namespaces work
- [ ] Cluster-scoped resources work

## Next Steps

After successful testing:

1. **Document any issues** you encounter
2. **Test with different RBAC configurations**
3. **Test with different resource types**
4. **Test error scenarios**
5. **Provide feedback** to improve the implementation

## Support

If you encounter issues during testing:

1. Check the troubleshooting section above
2. Review server logs (stderr output)
3. Verify Kubernetes cluster access with kubectl
4. Check ChatGPT logs and MCP configuration
5. Open an issue on GitHub with:
   - Error messages
   - Steps to reproduce
   - Kubernetes version
   - RBAC configuration (if applicable)

