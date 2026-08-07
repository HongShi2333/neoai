#!/bin/bash
# Full integration test for NeoAI — runs every feature end-to-end.
set +e

export PATH=/home/z/go/bin:$PATH
cd /home/z/my-project/coai

# Kill any leftover server
pkill -f "go run" 2>/dev/null
pkill -f "exe/neoai" 2>/dev/null
sleep 1

# Clean state
rm -rf db/chatnio.db config/config.yaml 2>/dev/null
mkdir -p /tmp/neoai-test

cat > config/config.yaml << 'EOF'
mysql:
  host: ""
cache:
  type: valkey
  host: ""
  port: 6379
secret: "0123456789012345678901234567890123456789012345678901234567890123"
serve_static: false
server:
  port: 18094
allow_origins:
  - http://localhost:5173
  - https://app.example.com

channel:
  - id: 1
    name: openai-public
    type: openai
    priority: 0
    weight: 1
    retry: 3
    secret: sk-test-public
    endpoint: https://api.openai.com
    mapper: ""
    state: true
    group: []
    models: [gpt-3.5-turbo, gpt-4o-mini]
    proxy: {proxytype: 0, proxy: "", username: "", password: ""}
  - id: 2
    name: openai-pro
    type: openai
    priority: 0
    weight: 1
    retry: 3
    secret: sk-test-pro
    endpoint: https://api.openai.com
    mapper: ""
    state: true
    group: ["pro", "admin"]
    models: [gpt-4, gpt-4-turbo, gpt-4o]
    proxy: {proxytype: 0, proxy: "", username: "", password: ""}

market:
  - {id: gpt-3.5-turbo, name: "GPT-3.5 Turbo", description: "OpenAI GPT-3.5 Turbo", default: true, highcontext: false, avatar: "", tag: []}
  - {id: gpt-4o-mini, name: "GPT-4o mini", description: "OpenAI GPT-4o mini", default: true, highcontext: false, avatar: "", tag: []}
  - {id: gpt-4, name: "GPT-4", description: "OpenAI GPT-4", default: false, highcontext: false, avatar: "", tag: []}
  - {id: gpt-4-turbo, name: "GPT-4 Turbo", description: "OpenAI GPT-4 Turbo", default: false, highcontext: false, avatar: "", tag: []}
  - {id: gpt-4o, name: "GPT-4o", description: "OpenAI GPT-4o", default: false, highcontext: false, avatar: "", tag: []}

charge:
  - id: 1
    type: token
    models: [gpt-3.5-turbo, gpt-4o-mini]
    input: 0.001
    output: 0.002
    anonymous: true
  - id: 2
    type: token
    models: [gpt-4, gpt-4-turbo, gpt-4o]
    input: 0.03
    output: 0.06
    anonymous: false
EOF

# Start server
nohup go run . > /tmp/neoai-test/server.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Wait for server
echo "Waiting for server..."
for i in {1..30}; do
  if curl -sS http://localhost:18094/healthz > /dev/null 2>&1; then
    echo "Server ready after ${i}s"
    break
  fi
  sleep 1
done

BASE=http://localhost:18094/api
PASS=0
FAIL=0
FAILED_TESTS=()

check() {
  local name="$1"
  local result="$2"
  if [ "$result" = "PASS" ]; then
    PASS=$((PASS+1))
    echo "  ✓ $name"
  else
    FAIL=$((FAIL+1))
    FAILED_TESTS+=("$name")
    echo "  ✗ $name — $result"
  fi
}

echo ""
echo "=========================================="
echo "  AUTH MODULE"
echo "=========================================="

LOGIN=$(curl -sS -X POST $BASE/login -H "Content-Type: application/json" -d '{"username":"root","password":"chatnio123456"}')
ADMIN_TOKEN=$(echo "$LOGIN" | python3 -c "import json,sys; print(json.load(sys.stdin).get('token',''))" 2>/dev/null)
if [ -n "$ADMIN_TOKEN" ]; then check "2.1 Login as root" "PASS"; else check "2.1 Login as root" "FAIL: $LOGIN"; fi

STATE=$(curl -sS -X POST $BASE/state -H "Authorization: $ADMIN_TOKEN")
echo "$STATE" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('admin') else 1)" && check "2.2 /state verifies token" "PASS" || check "2.2 /state verifies token" "FAIL: $STATE"

USERINFO=$(curl -sS $BASE/userinfo -H "Authorization: $ADMIN_TOKEN")
echo "$USERINFO" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('data',{}).get('email') == 'root@example.com' else 1)" && check "2.3 /userinfo" "PASS" || check "2.3 /userinfo" "FAIL: $USERINFO"

GEN=$(curl -sS -X POST $BASE/admin/registration-code/generate -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"number":2,"quota":100,"max_uses":3,"note":"test"}')
REG_CODE=$(echo "$GEN" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['data'][0] if d.get('status') and d.get('data') else '')" 2>/dev/null)
if [ -n "$REG_CODE" ]; then check "2.4 Generate registration codes" "PASS"; else check "2.4 Generate registration codes" "FAIL: $GEN"; fi

LIST=$(curl -sS $BASE/admin/registration-code/list -H "Authorization: $ADMIN_TOKEN")
echo "$LIST" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and len(d.get('data',[])) >= 2 else 1)" && check "2.5 List registration codes" "PASS" || check "2.5 List registration codes" "FAIL: $LIST"

STATERESP=$(curl -sS -X POST $BASE/admin/registration-code/state -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"required":true}')
echo "$STATERESP" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('required') == True else 1)" && check "2.6 Enable registration code requirement" "PASS" || check "2.6 Enable registration code requirement" "FAIL: $STATERESP"

NOCODE=$(curl -sS -X POST $BASE/register -H "Content-Type: application/json" -d '{"username":"test_nocode","password":"password123","email":"nocode@test.com"}')
echo "$NOCODE" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if not d.get('status') and 'required' in d.get('error','').lower() else 1)" && check "2.7 Register without code rejected" "PASS" || check "2.7 Register without code rejected" "FAIL: $NOCODE"

WRONGCODE=$(curl -sS -X POST $BASE/register -H "Content-Type: application/json" -d '{"username":"test_wrong","password":"password123","email":"wrong@test.com","registration_code":"REG-WRONG123"}')
echo "$WRONGCODE" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if not d.get('status') and 'invalid' in d.get('error','').lower() else 1)" && check "2.8 Register with wrong code rejected" "PASS" || check "2.8 Register with wrong code rejected" "FAIL: $WRONGCODE"

REGRESULT=$(curl -sS -X POST $BASE/register -H "Content-Type: application/json" -d "{\"username\":\"testuser1\",\"password\":\"password123\",\"email\":\"test1@test.com\",\"registration_code\":\"$REG_CODE\"}")
USER1_TOKEN=$(echo "$REGRESULT" | python3 -c "import json,sys; print(json.load(sys.stdin).get('token',''))" 2>/dev/null)
if [ -n "$USER1_TOKEN" ]; then check "2.9 Register with valid code" "PASS"; else check "2.9 Register with valid code" "FAIL: $REGRESULT"; fi

LOGIN2=$(curl -sS -X POST $BASE/login -H "Content-Type: application/json" -d '{"username":"testuser1","password":"password123"}')
echo "$LOGIN2" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('token') else 1)" && check "2.10 Login as new user" "PASS" || check "2.10 Login as new user" "FAIL: $LOGIN2"

RENAME=$(curl -sS -X POST $BASE/profile/username -H "Authorization: $USER1_TOKEN" -H "Content-Type: application/json" -d '{"username":"testuser1_renamed"}')
echo "$RENAME" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('username') == 'testuser1_renamed' else 1)" && check "2.11 Self-service username change" "PASS" || check "2.11 Self-service username change" "FAIL: $RENAME"

# Re-login as renamed user to get a fresh token (old token has stale username)
USER1_TOKEN=$(curl -sS -X POST $BASE/login -H "Content-Type: application/json" -d '{"username":"testuser1_renamed","password":"password123"}' | python3 -c "import json,sys; print(json.load(sys.stdin).get('token',''))" 2>/dev/null)

LOGIN3=$(curl -sS -X POST $BASE/login -H "Content-Type: application/json" -d '{"username":"testuser1_renamed","password":"password123"}')
echo "$LOGIN3" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('token') else 1)" && check "2.12 Login with new username" "PASS" || check "2.12 Login with new username" "FAIL: $LOGIN3"

ADMINRENAME=$(curl -sS -X POST $BASE/admin/user/username -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"id":2,"username":"admin_renamed_user"}')
echo "$ADMINRENAME" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "2.13 Admin renames user" "PASS" || check "2.13 Admin renames user" "FAIL: $ADMINRENAME"

# Re-login as the admin-renamed user (id=2 was testuser1_renamed,
# now renamed to admin_renamed_user) so USER1_TOKEN is fresh again.
USER1_TOKEN=$(curl -sS -X POST $BASE/login -H "Content-Type: application/json" -d '{"username":"admin_renamed_user","password":"password123"}' | python3 -c "import json,sys; print(json.load(sys.stdin).get('token',''))" 2>/dev/null)

OAUTHSAVE=$(curl -sS -X POST $BASE/admin/oauth/config -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"linuxdo":{"client_id":"ld_test","client_secret":"ld_sec"},"github":{"client_id":"gh_test","client_secret":"gh_sec"},"frontend_url":"https://app.example.com"}')
echo "$OAUTHSAVE" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "2.14 Set OAuth config" "PASS" || check "2.14 Set OAuth config" "FAIL: $OAUTHSAVE"

INFO=$(curl -sS $BASE/info)
echo "$INFO" | python3 -c "import json,sys; d=json.load(sys.stdin); ok = d.get('linuxdo_oauth_enabled') == True and d.get('github_oauth_enabled') == True and d.get('registration_code') == True; sys.exit(0 if ok else 1)" && check "2.15 /info exposes OAuth + reg-code flags" "PASS" || check "2.15 /info exposes OAuth + reg-code flags" "FAIL: $INFO"

APIKEY=$(curl -sS $BASE/apikey -H "Authorization: $USER1_TOKEN")
echo "$APIKEY" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('key') else 1)" && check "2.16 Get API key" "PASS" || check "2.16 Get API key" "FAIL: $APIKEY"

echo ""
echo "=========================================="
echo "  COMMUNITY CHANNELS"
echo "=========================================="

CHAN1=$(curl -sS -X POST $BASE/community/channels -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"name":"general","topic":"public chat","visibility":"public","send_permission":"everyone"}')
CHAN1_ID=$(echo "$CHAN1" | python3 -c "import json,sys; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [ -n "$CHAN1_ID" ]; then check "3.1 Create public channel" "PASS"; else check "3.1 Create public channel" "FAIL: $CHAN1"; fi

CHAN2=$(curl -sS -X POST $BASE/community/channels -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"name":"members-only","topic":"members only","visibility":"members","send_permission":"everyone"}')
CHAN2_ID=$(echo "$CHAN2" | python3 -c "import json,sys; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [ -n "$CHAN2_ID" ]; then check "3.2 Create members-only channel" "PASS"; else check "3.2 Create members-only channel" "FAIL: $CHAN2"; fi

CHAN3=$(curl -sS -X POST $BASE/community/channels -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"name":"announce","topic":"admin announcements","visibility":"public","send_permission":"admins"}')
CHAN3_ID=$(echo "$CHAN3" | python3 -c "import json,sys; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
if [ -n "$CHAN3_ID" ]; then check "3.3 Create admin-only-send channel" "PASS"; else check "3.3 Create admin-only-send channel" "FAIL: $CHAN3"; fi

LISTCHAN=$(curl -sS "$BASE/community/channels?all=1" -H "Authorization: $ADMIN_TOKEN")
echo "$LISTCHAN" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and len(d.get('data',[])) >= 3 else 1)" && check "3.4 List channels (admin, all)" "PASS" || check "3.4 List channels (admin, all)" "FAIL: $LISTCHAN"

LISTCHANUSER=$(curl -sS "$BASE/community/channels" -H "Authorization: $USER1_TOKEN")
echo "$LISTCHANUSER" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and len(d.get('data',[])) >= 3 else 1)" && check "3.5 List channels (user)" "PASS" || check "3.5 List channels (user)" "FAIL: $LISTCHANUSER"

MSG1=$(curl -sS -X POST "$BASE/community/channels/$CHAN1_ID/messages" -H "Authorization: $USER1_TOKEN" -H "Content-Type: application/json" -d '{"content":"hello from user"}')
echo "$MSG1" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('data',{}).get('content') == 'hello from user' else 1)" && check "3.6 User posts to public channel" "PASS" || check "3.6 User posts to public channel" "FAIL: $MSG1"

MSG2=$(curl -sS -X POST "$BASE/community/channels/$CHAN3_ID/messages" -H "Authorization: $USER1_TOKEN" -H "Content-Type: application/json" -d '{"content":"should fail"}')
echo "$MSG2" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if not d.get('status') and 'forbidden' in d.get('error','').lower() else 1)" && check "3.7 User cannot post to admin-only channel" "PASS" || check "3.7 User cannot post to admin-only channel" "FAIL: $MSG2"

MSG3=$(curl -sS -X POST "$BASE/community/channels/$CHAN3_ID/messages" -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"content":"admin announcement"}')
echo "$MSG3" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "3.8 Admin posts to admin-only channel" "PASS" || check "3.8 Admin posts to admin-only channel" "FAIL: $MSG3"

MSGS=$(curl -sS "$BASE/community/channels/$CHAN1_ID/messages" -H "Authorization: $USER1_TOKEN")
echo "$MSGS" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and len(d.get('data',[])) >= 1 else 1)" && check "3.9 List messages" "PASS" || check "3.9 List messages" "FAIL: $MSGS"

MSG_ID=$(echo "$MSG1" | python3 -c "import json,sys; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null)
DELMSG=$(curl -sS -X DELETE "$BASE/community/messages/$MSG_ID" -H "Authorization: $USER1_TOKEN")
echo "$DELMSG" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "3.10 Delete own message" "PASS" || check "3.10 Delete own message" "FAIL: $DELMSG"

UPDATECHAN=$(curl -sS -X POST "$BASE/community/channels/$CHAN1_ID" -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"name":"general-renamed","topic":"updated topic","visibility":"public","send_permission":"everyone"}')
echo "$UPDATECHAN" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('data',{}).get('name') == 'general-renamed' else 1)" && check "3.11 Update channel" "PASS" || check "3.11 Update channel" "FAIL: $UPDATECHAN"

DELCHAN=$(curl -sS -X DELETE "$BASE/community/channels/$CHAN3_ID" -H "Authorization: $ADMIN_TOKEN")
echo "$DELCHAN" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "3.12 Delete channel" "PASS" || check "3.12 Delete channel" "FAIL: $DELCHAN"

echo ""
echo "=========================================="
echo "  CHANNEL & PRICING"
echo "=========================================="

CHANS=$(curl -sS $BASE/admin/channel/list -H "Authorization: $ADMIN_TOKEN")
echo "$CHANS" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and len(d.get('data',[])) >= 2 else 1)" && check "4.1 List channels" "PASS" || check "4.1 List channels" "FAIL: $CHANS"

CHARGES=$(curl -sS $BASE/admin/charge/list -H "Authorization: $ADMIN_TOKEN")
echo "$CHARGES" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and len(d.get('data',[])) >= 2 else 1)" && check "4.2 List charge rules" "PASS" || check "4.2 List charge rules" "FAIL: $CHARGES"

JSONEXPORT=$(curl -sS $BASE/admin/charge/json -H "Authorization: $ADMIN_TOKEN")
echo "$JSONEXPORT" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and len(d.get('data',[])) >= 2 else 1)" && check "4.3 Export charge JSON" "PASS" || check "4.3 Export charge JSON" "FAIL: $JSONEXPORT"

JSONIMPORT=$(curl -sS -X POST "$BASE/admin/charge/json?mode=merge" -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '[{"type":"token","models":["test-model-1","test-model-2"],"input":0.5,"output":1.0,"anonymous":false}]')
echo "$JSONIMPORT" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "4.4 Import charge JSON (merge)" "PASS" || check "4.4 Import charge JSON (merge)" "FAIL: $JSONIMPORT"

CHARGES2=$(curl -sS $BASE/admin/charge/list -H "Authorization: $ADMIN_TOKEN")
echo "$CHARGES2" | python3 -c "import json,sys; d=json.load(sys.stdin); data = d.get('data',[]); has_test = any('test-model-1' in c.get('models',[]) for c in data); sys.exit(0 if has_test else 1)" && check "4.5 Imported charge rule exists" "PASS" || check "4.5 Imported charge rule exists" "FAIL: $CHARGES2"

FETCH=$(curl -sS -X POST $BASE/admin/channel/fetch-models -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"type":"openai","endpoint":"https://api.openai.com","secret":"sk-fake"}')
echo "$FETCH" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if 'status' in d else 1)" && check "4.6 Fetch models from upstream (graceful failure)" "PASS" || check "4.6 Fetch models from upstream" "FAIL: $FETCH"

echo ""
echo "=========================================="
echo "  MODEL LIST (GROUP FILTERING)"
echo "=========================================="

ADMINMODELS=$(curl -sS $BASE/v1/models -H "Authorization: $ADMIN_TOKEN")
echo "$ADMINMODELS" | python3 -c "
import json,sys
d=json.load(sys.stdin)
models = [m['id'] for m in d.get('data',[])]
expected = {'gpt-3.5-turbo','gpt-4o-mini','gpt-4','gpt-4-turbo','gpt-4o'}
actual = set(models)
sys.exit(0 if expected.issubset(actual) else 1)
" && check "5.1 Admin sees all models" "PASS" || check "5.1 Admin sees all models" "FAIL: $ADMINMODELS"

ADMINMARKET=$(curl -sS $BASE/v1/market -H "Authorization: $ADMIN_TOKEN")
echo "$ADMINMARKET" | python3 -c "
import json,sys
d=json.load(sys.stdin)
models = [m['id'] for m in d]
expected = {'gpt-3.5-turbo','gpt-4o-mini','gpt-4','gpt-4-turbo','gpt-4o'}
actual = set(models)
sys.exit(0 if expected.issubset(actual) else 1)
" && check "5.2 Admin sees all market entries" "PASS" || check "5.2 Admin sees all market entries" "FAIL: $ADMINMARKET"

USERMODELS=$(curl -sS $BASE/v1/models -H "Authorization: $USER1_TOKEN")
echo "$USERMODELS" | python3 -c "
import json,sys
d=json.load(sys.stdin)
models = set(m['id'] for m in d.get('data',[]))
pro_models = {'gpt-4','gpt-4-turbo','gpt-4o'}
disallowed = models & pro_models
sys.exit(0 if not disallowed else 1)
" && check "5.3 Normal user does NOT see pro models" "PASS" || check "5.3 Normal user does NOT see pro models" "FAIL: $USERMODELS"

echo "$USERMODELS" | python3 -c "
import json,sys
d=json.load(sys.stdin)
models = set(m['id'] for m in d.get('data',[]))
public_models = {'gpt-3.5-turbo','gpt-4o-mini'}
sys.exit(0 if public_models.issubset(models) else 1)
" && check "5.4 Normal user sees public models" "PASS" || check "5.4 Normal user sees public models" "FAIL: $USERMODELS"

ANONMODELS=$(curl -sS $BASE/v1/models)
echo "$ANONMODELS" | python3 -c "
import json,sys
d=json.load(sys.stdin)
models = set(m['id'] for m in d.get('data',[]))
sys.exit(0 if 'gpt-3.5-turbo' in models and 'gpt-4' not in models else 1)
" && check "5.5 Anonymous sees only anonymous-allowed models" "PASS" || check "5.5 Anonymous sees only anonymous-allowed models" "FAIL: $ANONMODELS"

echo ""
echo "=========================================="
echo "  USER MANAGEMENT"
echo "=========================================="

USERS=$(curl -sS "$BASE/admin/user/list" -H "Authorization: $ADMIN_TOKEN")
echo "$USERS" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('total',0) >= 1 else 1)" && check "6.1 List users" "PASS" || check "6.1 List users" "FAIL: $USERS"

QUOTA=$(curl -sS -X POST $BASE/admin/user/quota -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"id":2,"quota":50,"override":true}')
echo "$QUOTA" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "6.2 Set user quota" "PASS" || check "6.2 Set user quota" "FAIL: $QUOTA"

BAN=$(curl -sS -X POST $BASE/admin/user/ban -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"id":2,"ban":true}')
echo "$BAN" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "6.3 Ban user" "PASS" || check "6.3 Ban user" "FAIL: $BAN"

BANNEDLOGIN=$(curl -sS -X POST $BASE/login -H "Content-Type: application/json" -d '{"username":"admin_renamed_user","password":"password123"}')
echo "$BANNEDLOGIN" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if not d.get('status') and 'banned' in d.get('error','').lower() else 1)" && check "6.4 Banned user cannot login" "PASS" || check "6.4 Banned user cannot login" "FAIL: $BANNEDLOGIN"

UNBAN=$(curl -sS -X POST $BASE/admin/user/ban -H "Authorization: $ADMIN_TOKEN" -H "Content-Type: application/json" -d '{"id":2,"ban":false}')
echo "$UNBAN" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') else 1)" && check "6.5 Unban user" "PASS" || check "6.5 Unban user" "FAIL: $UNBAN"

echo ""
echo "=========================================="
echo "  HEALTH & SYSTEM"
echo "=========================================="

HEALTH=$(curl -sS http://localhost:18094/healthz)
echo "$HEALTH" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') == 'ok' else 1)" && check "7.1 /healthz" "PASS" || check "7.1 /healthz" "FAIL: $HEALTH"

READY=$(curl -sS http://localhost:18094/ready)
echo "$READY" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('db') == 'up' else 1)" && check "7.2 /ready" "PASS" || check "7.2 /ready" "FAIL: $READY"

INFO2=$(curl -sS $BASE/info)
echo "$INFO2" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if 'title' in d else 1)" && check "7.3 /info public" "PASS" || check "7.3 /info public" "FAIL: $INFO2"

MODELSFORMAT=$(curl -sS $BASE/v1/models -H "Authorization: $ADMIN_TOKEN")
echo "$MODELSFORMAT" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('object') == 'list' and isinstance(d.get('data'),list) else 1)" && check "7.4 /v1/models OpenAI-compatible format" "PASS" || check "7.4 /v1/models OpenAI-compatible format" "FAIL: $MODELSFORMAT"

CONFIG=$(curl -sS $BASE/admin/config/view -H "Authorization: $ADMIN_TOKEN")
echo "$CONFIG" | python3 -c "import json,sys; d=json.load(sys.stdin); sys.exit(0 if d.get('status') and d.get('data') else 1)" && check "7.5 Admin config view" "PASS" || check "7.5 Admin config view" "FAIL: $CONFIG"

CORS=$(curl -sS -I -X OPTIONS $BASE/login -H "Origin: https://app.example.com" -H "Access-Control-Request-Method: POST" 2>&1)
echo "$CORS" | grep -i "access-control-allow-origin" > /dev/null && check "7.6 CORS headers present" "PASS" || check "7.6 CORS headers present" "FAIL: no CORS header"

echo ""
echo "=========================================="
echo "  SUMMARY"
echo "=========================================="
echo "  PASS: $PASS"
echo "  FAIL: $FAIL"
if [ $FAIL -gt 0 ]; then
  echo ""
  echo "  Failed tests:"
  for t in "${FAILED_TESTS[@]}"; do
    echo "    - $t"
  done
fi
echo "=========================================="

kill $SERVER_PID 2>/dev/null
pkill -f "go run" 2>/dev/null
pkill -f "exe/neoai" 2>/dev/null

exit $FAIL
