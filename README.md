# Write Me

자기소개서 작성 보조 서비스의 MVP 구현이다.

## 구성

- `apps/web`: Next.js 웹 앱
- `apps/api`: Go API 서버
- `db`: PostgreSQL
- `scripts/bridge`: 호스트 네이티브 LLM bridge 설치/실행 스크립트

## 빠른 시작

1. `.env.example`을 `.env`로 복사하고 값을 채운다.
2. 브리지를 설치한다.
   - `make bridge-install`
3. 브리지를 실행하고 상태를 확인한다.
   - `make bridge-start`
   - `make bridge-doctor`
   - provider 로그인 필요 시 `./scripts/bridge/auth.sh codex|gemini|claude`
4. 앱 스택을 올린다.
   - `make up`
5. 브라우저에서 `http://localhost:3000`에 접속한다.

## 기본 관리자 계정

- 이메일: `.env`의 `ADMIN_EMAIL`
- 비밀번호: `.env`의 `ADMIN_PASSWORD`

## 브리지 메모

- MVP 기본 브리지는 `CLIProxyAPI`
- 브리지는 `127.0.0.1:${BRIDGE_PORT:-43110}`에만 바인딩된다.
- API 컨테이너는 `host.docker.internal`을 통해 브리지에 연결된다.
- 브리지 점검은 `/v1/models` 기준으로 한다. provider 로그인 전에는 응답이 `{"data":[]}`일 수 있다.

## 주요 경로

- 웹: `http://localhost:3000`
- API: `http://localhost:8080`
- 브리지 models: `http://127.0.0.1:43110/v1/models`
