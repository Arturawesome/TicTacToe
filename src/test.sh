#!/usr/bin/env bash
set -u
set -o pipefail

BASE_URL="${BASE_URL:-http://localhost:4000}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

PASS=0
FAIL=0
STATUS=""
BODY=""
TS="$(date +%s)"
A_LOGIN="jwt_alice_${TS}"
B_LOGIN="jwt_bob_${TS}"
PASSWORD="123456"

pass() { PASS=$((PASS+1)); echo "PASS: $1"; }
fail() { FAIL=$((FAIL+1)); echo "FAIL: $1"; echo "      HTTP=$STATUS"; echo "      BODY=$BODY"; }

request() {
  local method="$1" path="$2" token="${3:-}" kind="${4:-}" data="${5:-}"
  local body_file="$TMP_DIR/body"
  local args=(-sS -X "$method" -o "$body_file" -w '%{http_code}')

  if [[ -n "$token" ]]; then
    args+=(-H "Authorization: Bearer $token")
  fi
  if [[ "$kind" == "json" ]]; then
    args+=(-H 'Content-Type: application/json')
  elif [[ "$kind" == "form" ]]; then
    args+=(-H 'Content-Type: application/x-www-form-urlencoded')
  fi
  if [[ -n "$data" ]]; then
    args+=(--data "$data")
  fi

  STATUS="$(curl "${args[@]}" "$BASE_URL$path" 2>"$TMP_DIR/curl.err" || true)"
  BODY="$(cat "$body_file" 2>/dev/null || true)"
}

expect_status() {
  local expected="$1" msg="$2"
  if [[ "$STATUS" == "$expected" ]]; then pass "$msg (HTTP $STATUS)"; else fail "$msg; expected $expected"; fi
}

json_value() {
  local key="$1"
  printf '%s' "$BODY" | python3 -c 'import json,sys; d=json.load(sys.stdin); print(d[sys.argv[1]])' "$key" 2>/dev/null
}

register_user() {
  local login="$1"
  request POST /user/register "" form "login=${login}&password=${PASSWORD}"
  expect_status 201 "register $login"
}

login_user() {
  local login="$1"
  request POST /user/login "" json "{\"login\":\"${login}\",\"password\":\"${PASSWORD}\"}"
  expect_status 200 "login $login"
  if [[ "$STATUS" == "200" ]]; then
    local token
    token="$(json_value accessToken)"
    if [[ -n "$token" ]]; then pass "$login received accessToken"; else fail "$login did not receive accessToken"; fi
    printf '%s' "$token"
  fi
}

create_game() {
  local token="$1"
  request POST /games "$token" form 'opponent=0'
  expect_status 201 "create human-vs-human game"
  GAME_ID=""
  if [[ "$STATUS" == "201" ]]; then
    GAME_ID="$(json_value id)"
    [[ -n "$GAME_ID" ]] && pass "game id: $GAME_ID" || fail "game id missing"
  fi
}

join_game() {
  local token="$1" gid="$2"
  request POST "/games/${gid}/join" "$token"
  expect_status 200 "second player joins $gid"
}

move() {
  local token="$1" gid="$2" i="$3" j="$4" label="$5"
  request POST "/games/${gid}/move" "$token" form "i=${i}&j=${j}"
  expect_status 200 "$label"
}

assert_finished() {
  local token="$1" gid="$2" expected_winner="$3" label="$4"
  request GET "/games/${gid}" "$token"
  expect_status 200 "fetch finished game $gid"
  if [[ "$STATUS" == "200" ]]; then
    if printf '%s' "$BODY" | python3 -c '
import json,sys
x=json.load(sys.stdin)
expected=sys.argv[1]
assert x["GameStatus"] in (2,3)
if expected == "DRAW":
    assert x["GameStatus"] == 2
    assert x.get("Winner") is None
else:
    assert x["GameStatus"] == 3
    assert x.get("Winner") == expected
' "$expected_winner" 2>/dev/null; then
      pass "$label"
    else
      fail "$label"
    fi
  fi
}

echo "============================================================"
echo "JWT + TWO-PLAYER GAMES + HISTORY + LEADERBOARD"
echo "BASE_URL=$BASE_URL"
echo "============================================================"

echo
echo "[1] Protected endpoints require Bearer JWT"
request GET /games/history
expect_status 401 "history rejects unauthenticated request"
request GET '/games/leaderboard?n=10'
expect_status 401 "leaderboard rejects unauthenticated request"

echo
echo "[2] Register and authorize two users"
register_user "$A_LOGIN"
register_user "$B_LOGIN"

request POST /user/login "" json "{\"login\":\"${A_LOGIN}\",\"password\":\"${PASSWORD}\"}"
expect_status 200 "Alice JWT login"
A_ACCESS="$(json_value accessToken)"
A_REFRESH="$(json_value refreshToken)"
A_TYPE="$(json_value type)"
[[ "$A_TYPE" == "Bearer" && -n "$A_ACCESS" && -n "$A_REFRESH" ]] && pass "Alice JwtResponse has Bearer/access/refresh" || fail "Alice JwtResponse is invalid"

request POST /user/login "" json "{\"login\":\"${B_LOGIN}\",\"password\":\"${PASSWORD}\"}"
expect_status 200 "Bob JWT login"
B_ACCESS="$(json_value accessToken)"
B_REFRESH="$(json_value refreshToken)"

request GET /user/me "$A_ACCESS"
expect_status 200 "Alice /user/me by accessToken"
A_ID="$(json_value id)"
A_ME_LOGIN="$(json_value login)"
[[ "$A_ME_LOGIN" == "$A_LOGIN" ]] && pass "Alice /me login matches" || fail "Alice /me login mismatch"

request GET /user/me "$B_ACCESS"
expect_status 200 "Bob /user/me by accessToken"
B_ID="$(json_value id)"
B_ME_LOGIN="$(json_value login)"
[[ "$B_ME_LOGIN" == "$B_LOGIN" ]] && pass "Bob /me login matches" || fail "Bob /me login mismatch"

request GET /user/me "$A_ACCESS"  # request() always adds Bearer
expect_status 200 "Bearer access token works"

# Wrong format: raw token without Bearer
RAW_BODY_FILE="$TMP_DIR/raw_body"
RAW_STATUS="$(curl -sS -o "$RAW_BODY_FILE" -w '%{http_code}' -H "Authorization: $A_ACCESS" "$BASE_URL/user/me" || true)"
if [[ "$RAW_STATUS" == "401" ]]; then pass "raw token without Bearer is rejected"; else STATUS="$RAW_STATUS"; BODY="$(cat "$RAW_BODY_FILE")"; fail "raw token without Bearer must be 401"; fi

echo
echo "[3] Refresh accessToken without Authorization header"
request POST /user/refresh/accessToken "" json "{\"refreshToken\":\"${A_REFRESH}\"}"
expect_status 200 "accessToken refresh is public and succeeds"
if [[ "$STATUS" == "200" ]]; then
  NEW_A_ACCESS="$(json_value accessToken)"
  RETURNED_REFRESH="$(json_value refreshToken)"
  [[ -n "$NEW_A_ACCESS" && "$RETURNED_REFRESH" == "$A_REFRESH" ]] && pass "refresh access returns accessToken and keeps refreshToken" || fail "refresh access response invalid"
  A_ACCESS="$NEW_A_ACCESS"
fi

request GET /user/me "$A_ACCESS"
expect_status 200 "refreshed accessToken authorizes /user/me"

echo
echo "[4] Game 1: Alice beats Bob"
create_game "$A_ACCESS"; G1="$GAME_ID"; join_game "$B_ACCESS" "$G1"
move "$A_ACCESS" "$G1" 0 0 "G1 Alice X (0,0)"
move "$B_ACCESS" "$G1" 1 0 "G1 Bob O (1,0)"
move "$A_ACCESS" "$G1" 0 1 "G1 Alice X (0,1)"
move "$B_ACCESS" "$G1" 1 1 "G1 Bob O (1,1)"
move "$A_ACCESS" "$G1" 0 2 "G1 Alice X wins"
assert_finished "$A_ACCESS" "$G1" "$A_ID" "G1 winner UUID is Alice"

echo
echo "[5] Game 2: Bob beats Alice"
create_game "$B_ACCESS"; G2="$GAME_ID"; join_game "$A_ACCESS" "$G2"
move "$B_ACCESS" "$G2" 0 0 "G2 Bob X (0,0)"
move "$A_ACCESS" "$G2" 1 0 "G2 Alice O (1,0)"
move "$B_ACCESS" "$G2" 0 1 "G2 Bob X (0,1)"
move "$A_ACCESS" "$G2" 1 1 "G2 Alice O (1,1)"
move "$B_ACCESS" "$G2" 0 2 "G2 Bob X wins"
assert_finished "$B_ACCESS" "$G2" "$B_ID" "G2 winner UUID is Bob"

echo
echo "[6] Game 3: Alice beats Bob again"
create_game "$A_ACCESS"; G3="$GAME_ID"; join_game "$B_ACCESS" "$G3"
move "$A_ACCESS" "$G3" 0 0 "G3 Alice X (0,0)"
move "$B_ACCESS" "$G3" 1 0 "G3 Bob O (1,0)"
move "$A_ACCESS" "$G3" 0 1 "G3 Alice X (0,1)"
move "$B_ACCESS" "$G3" 1 1 "G3 Bob O (1,1)"
move "$A_ACCESS" "$G3" 0 2 "G3 Alice X wins"
assert_finished "$A_ACCESS" "$G3" "$A_ID" "G3 winner UUID is Alice"

echo
echo "[7] Game 4: draw"
create_game "$A_ACCESS"; G4="$GAME_ID"; join_game "$B_ACCESS" "$G4"
# Final board: X O X / X O O / O X X
move "$A_ACCESS" "$G4" 0 0 "G4 X (0,0)"
move "$B_ACCESS" "$G4" 0 1 "G4 O (0,1)"
move "$A_ACCESS" "$G4" 0 2 "G4 X (0,2)"
move "$B_ACCESS" "$G4" 1 1 "G4 O (1,1)"
move "$A_ACCESS" "$G4" 1 0 "G4 X (1,0)"
move "$B_ACCESS" "$G4" 1 2 "G4 O (1,2)"
move "$A_ACCESS" "$G4" 2 1 "G4 X (2,1)"
move "$B_ACCESS" "$G4" 2 0 "G4 O (2,0)"
move "$A_ACCESS" "$G4" 2 2 "G4 final X -> draw"
assert_finished "$A_ACCESS" "$G4" DRAW "G4 status is Draw and Winner is null"

echo
echo "[8] History must contain ALL completed games, including losses"
request GET /games/history "$A_ACCESS"
expect_status 200 "Alice history endpoint"
if [[ "$STATUS" == "200" ]]; then
  if printf '%s' "$BODY" | python3 -c '
import json,sys
ids=set(sys.argv[1:])
data=json.load(sys.stdin)
actual={g["ID"] for g in data}
assert ids.issubset(actual), (ids,actual)
for g in data:
    if g["ID"] in ids:
        assert g["GameStatus"] in (2,3)
        assert g.get("CreatedAt")
' "$G1" "$G2" "$G3" "$G4" 2>/dev/null; then
    pass "Alice history contains wins, loss, draw and CreatedAt"
  else
    fail "Alice history must contain all four completed games"
  fi
fi

request GET /games/history "$B_ACCESS"
expect_status 200 "Bob history endpoint"
if [[ "$STATUS" == "200" ]]; then
  if printf '%s' "$BODY" | python3 -c '
import json,sys
ids=set(sys.argv[1:])
data=json.load(sys.stdin)
actual={g["ID"] for g in data}
assert ids.issubset(actual), (ids,actual)
' "$G1" "$G2" "$G3" "$G4" 2>/dev/null; then
    pass "Bob history contains wins, losses and draw"
  else
    fail "Bob history must contain all four completed games"
  fi
fi

echo
echo "[9] Leaderboard"
request GET '/games/leaderboard?n=100' "$A_ACCESS"
expect_status 200 "leaderboard endpoint"
if [[ "$STATUS" == "200" ]]; then
  if printf '%s' "$BODY" | python3 -c '
import json,sys,math
alice,bob=sys.argv[1],sys.argv[2]
data=json.load(sys.stdin)
by_login={x["login"]:x for x in data}
a=by_login[alice]; b=by_login[bob]
assert a["user_id"]
assert b["user_id"]
# Alice: 2 wins / (1 loss + 1 draw) = 1.0
# Bob:   1 win  / (2 losses + 1 draw) = 1/3
assert math.isclose(float(a["win_ratio"]),1.0,rel_tol=1e-9,abs_tol=1e-9), a
assert math.isclose(float(b["win_ratio"]),1.0/3.0,rel_tol=1e-9,abs_tol=1e-9), b
assert float(a["win_ratio"]) > float(b["win_ratio"])
' "$A_LOGIN" "$B_LOGIN" 2>/dev/null; then
    pass "leaderboard ratios are correct: Alice=1.0, Bob=1/3"
  else
    fail "leaderboard does not contain expected users/ratios"
  fi
fi

request GET '/games/leaderboard?n=0' "$A_ACCESS"
expect_status 400 "leaderboard rejects n <= 0"
request GET '/games/leaderboard?n=abc' "$A_ACCESS"
expect_status 400 "leaderboard rejects non-numeric n"

echo
echo "============================================================"
echo "RESULT: PASS=$PASS FAIL=$FAIL"
echo "Users: $A_LOGIN ($A_ID), $B_LOGIN ($B_ID)"
echo "Games: $G1 $G2 $G3 $G4"
echo "============================================================"

if [[ "$FAIL" -ne 0 ]]; then
  exit 1
fi
