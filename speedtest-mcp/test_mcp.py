import subprocess, json

proc = subprocess.Popen(
    ["go", "run", "main.go"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
   # stderr=subprocess.PIPE
)

def send(msg):
    line = json.dumps(msg) + "\n"
    proc.stdin.write(line.encode())
    proc.stdin.flush()
    return json.loads(proc.stdout.readline())

# Step 1: initialize
resp = send({"jsonrpc":"2.0","id":1,"method":"initialize","params":{
    "protocolVersion":"2024-11-05",
    "capabilities":{},
    "clientInfo":{"name":"test","version":"1.0"}
}})
print("initialize:", resp)

# Step 2: initialized notification (no response expected)
proc.stdin.write((json.dumps({"jsonrpc":"2.0","method":"notifications/initialized","params":{}}) + "\n").encode())
proc.stdin.flush()

# Step 3: tools/list
resp = send({"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}})
print("tools/list:", resp)

resp = send({"jsonrpc":"2.0","id":3,"method":"tools/call","params": {
    "name" : "getAllResults",
    "arguments": {}
}})
print("getAllResults:", resp)
resp = send({"jsonrpc":"2.0","id":4,"method":"tools/call","params":{
    "name": "getSlowSpeedResults",
    "arguments": {"speedthreshold": 90}
}})
print("tools/call:", resp)

proc.stdin.close()