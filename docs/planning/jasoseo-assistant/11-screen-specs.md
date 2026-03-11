# 11. Screen Specs

## 공통 기준

### 공통 라우트 규칙
- 인증 전 진입 가능 경로: `/login`
- 인증 후 기본 진입 경로: `/dashboard`
- 핵심 화면은 모두 deep link 가능한 URL을 가진다.
- `:id`는 `WritingSession.id`를 의미한다.

### 공통 레이아웃 규칙
- 기준 해상도: desktop-first
- 데스크톱 그리드: 12 columns
- 앱 셸:
  - 좌측 내비게이션 rail: `88px`
  - 좌측 자료 패널: `280px`
  - 본문 작업영역 최소 폭: `640px`
  - 우측 AI 패널: `360px`
- 반응형 breakpoint:
  - `1440`: wide desktop
  - `1280`: default desktop
  - `1024`: compact desktop/tablet landscape
  - `768`: tablet portrait / mobile landscape
  - `390`: mobile portrait
- 수평 스크롤 금지

### 공통 상태 규칙
- 모든 화면은 `default / empty / loading / error / success`를 가진다.
- destructive action은 항상 확인 다이얼로그를 거친다.
- 모든 async action은 300ms 이상일 경우 loading feedback을 보여준다.

### 공통 단축키
- `?`: 단축키 도움말 열기
- `Esc`: 현재 열려 있는 dialog, sheet, panel 닫기
- `Cmd/Ctrl+S`: 즉시 저장
- `Cmd/Ctrl+/`: AI 채팅 패널 포커스

## Route: `/login`

### 목적
- 단일 관리자 계정으로 서비스 접근

### 진입 조건
- 인증되지 않은 상태
- 세션 만료 상태

### 이탈 조건
- 로그인 성공 시 `/dashboard`로 이동

### 정보 구조
- 브랜드/서비스 설명 패널
- 로그인 폼 패널
- 설치 환경 안내 텍스트

### 핵심 CTA
- `로그인`

### 필드 목록
- 이메일
  - type: email
  - required: true
- 비밀번호
  - type: password
  - required: true

### 검증 규칙
- 이메일은 유효한 형식이어야 한다.
- 비밀번호는 비어 있을 수 없다.
- 인증 실패 시 일반 오류 문구를 노출한다.

### 비정상 상태
- 인증 실패
- 네트워크 오류
- 저장된 세션 손상

### 단축키
- `Enter`: 로그인 제출

### 모바일 규칙
- 2컬럼을 1컬럼으로 축소
- 브랜드 패널은 폼 아래 접이식 요약으로 이동

## Route: `/dashboard`

### 목적
- 새 작성 시작점 제공
- 최근 세션과 자료 현황 요약

### 진입 조건
- 로그인 완료

### 이탈 조건
- 새 세션 생성
- 기존 세션 열기
- 자료함, 설정 이동

### 정보 구조
- 상단 페이지 헤더
- `새 작성 시작` 카드 3개
- 최근 작업 리스트
- 자료 라이브러리 요약
- 빠른 팁 / 시스템 상태

### 핵심 CTA
- `지원 페이지 붙여넣고 시작`
- `기존 초안 붙여넣고 시작`
- `자료 선택 후 시작`

### 필드 목록
- 최근 작업 검색어
- 최근 작업 필터
  - 전체 / 진행 중 / 제출 준비 완료

### 검증 규칙
- 검색어는 공백만 입력 불가

### 비정상 상태
- 최근 작업 로드 실패
- 자료 요약 로드 실패

### 단축키
- `N`: 새 세션 시작
- `/`: 최근 작업 검색 포커스

### 모바일 규칙
- 시작 카드는 세로 스택
- 최근 작업 리스트는 1열 카드

## Route: `/sessions/new`

### 목적
- 새 작성 세션 생성
- 지원 페이지/기존 초안/자료 기반 시작 방식을 하나의 흐름으로 제공

### 진입 조건
- 로그인 완료

### 이탈 조건
- 세션 생성 성공 시 `/sessions/:id` 이동
- 취소 시 `/dashboard` 복귀

### 정보 구조
- 헤더: `새 작성 시작`
- Step 1: 시작 방식 선택
- Step 2: 입력값 수집
- Step 3: 자동 구조화 결과 확인
- Step 4: 자료 연결
- 하단 sticky CTA bar

### 시작 방식
- `지원 페이지 붙여넣기`
- `기존 초안 붙여넣기`
- `자료 선택 후 시작`

### 핵심 CTA
- `구조화 실행`
- `작성 세션 만들기`

### 필드 목록
- 지원 페이지 텍스트
  - textarea
  - required: mode가 `지원 페이지 붙여넣기`일 때 true
- 기존 초안 텍스트
  - textarea
  - required: mode가 `기존 초안 붙여넣기`일 때 true
- 회사명
  - text
  - editable after parse
- 직무명
  - text
  - editable after parse
- 질문 목록
  - repeated fields
  - each item: 질문 제목, 질문 본문, 글자 수 제한
- 자료 선택
  - multi-select
  - required: mode가 `자료 선택 후 시작`일 때 1개 이상

### 검증 규칙
- 지원 페이지 텍스트는 100자 이상일 때 파싱 실행 허용
- 기존 초안은 30자 이상일 때 진행 가능
- 질문 글자 수 제한은 `1~5000` 범위 정수
- 회사명/직무명은 비어 있을 수 있으나, 비어 있으면 UI에 `미확정` badge 표시

### 비정상 상태
- 파싱 실패
- 일부 질문만 추출 성공
- 지원 포맷 미인식
- 세션 생성 실패

### 단축키
- `Cmd/Ctrl+Enter`: 구조화 실행
- `Cmd/Ctrl+S`: 현재 입력 임시 저장

### 모바일 규칙
- stepper를 상단 segmented progress로 축약
- 질문 목록 편집은 full-width sheet 사용

## Route: `/sessions/:id`

### 목적
- 세션의 메인 작성 화면
- 문항별 편집, 자료 확인, AI 채팅, 도구 실행, 자동 저장, 자동 검토 수행

### 진입 조건
- 유효한 세션 ID 존재

### 이탈 조건
- 비교 화면 이동
- 검토 화면 이동
- 대시보드 또는 다른 화면 이동

### 정보 구조
- 상단 context bar
  - 회사명
  - 직무명
  - 저장 상태
  - 글자 수 기준 토글
  - 자동 검토 토글
  - 자동 반영 토글
- 문항 탭 strip
- 좌측 자료 패널
- 중앙 에디터
- 우측 도구 메뉴 + AI 채팅
- quick review drawer
- 하단 sticky action bar

### 핵심 CTA
- `검토 열기`
- `비교 보기`
- `최종 확정`

### 필드 목록
- 글자 수 기준 토글
  - 공백 포함 / 공백 제외
- 자동 검토
  - default: ON
- 자동 반영
  - default: OFF
- 적용 범위
  - 선택 문단 / 전체 문서
- 질문 탭
  - repeated
- 본문 에디터
  - paragraph-based annotated editor
  - 출력은 plain text지만, 편집 중에는 문단 선택, inline `ClaimTag`, diff anchor를 함께 지원해야 한다.
- AI 채팅 입력창
  - multi-line

### 검증 규칙
- `최종 확정`은 unresolved `INFERRED`가 1개라도 있으면 disabled
- 글자 수 초과는 block하지 않지만 warning state 표시
- 자동 반영 ON일 때도 변경 전 snapshot이 먼저 생성돼야 한다.

### 비정상 상태
- 세션 로드 실패
- 저장 실패
- AI 응답 실패
- 검토 실패
- diff 생성 실패

### 단축키
- `Alt+1..9`: 질문 탭 전환
- `Cmd/Ctrl+R`: 검토 실행
- `Cmd/Ctrl+Shift+C`: 비교 화면 열기
- `Cmd/Ctrl+/`: AI 채팅 포커스

### 모바일 규칙
- 좌측 자료 패널은 bottom sheet
- 우측 AI 패널은 full-height sheet
- 하단 action bar는 2열 버튼 묶음으로 축소

### 추가 규칙
- `검토 열기`는 먼저 editor 안의 quick review drawer를 연다.
- `전체 리뷰 보기` 선택 시 `/sessions/:id/review`로 이동한다.
- 에디터는 단순 `textarea`로 해석하지 않는다. `INFERRED` 표시, 문단 범위 선택, 변경 제안 anchor를 동시에 다룰 수 있는 주석형 편집 영역이어야 한다.

## Route: `/sessions/:id/compare`

### 목적
- 기존안과 수정안을 병렬 비교하고 적용 여부를 결정

### 진입 조건
- 세션 안에 비교 대상 제안(diff suggestion)이 존재

### 이탈 조건
- 적용 후 `/sessions/:id` 복귀
- 취소 후 `/sessions/:id` 복귀

### 정보 구조
- 상단 meta bar
  - 대상 문항
  - 적용 범위
  - 생성 시각
  - 생성 출처(도구/채팅)
- 좌측 기존안
- 우측 수정안
- 하단 sticky action bar

### 핵심 CTA
- `이 수정안 적용`
- `적용 안 함`

### 필드 목록
- diff chunk navigation
- 제안 선택
  - 현재 제안 / 이전 / 다음

### 검증 규칙
- diff 데이터가 없으면 빈 상태와 `에디터로 돌아가기`만 노출하고 적용 CTA는 숨김

### 비정상 상태
- diff 로드 실패
- 적용 실패

### 단축키
- `A`: 현재 수정안 적용
- `X`: 현재 수정안 적용 안 함
- `J / K`: 변경 chunk 이동

### 모바일 규칙
- 상하 stacked 비교
- 변경 chunk 단위로 아코디언 노출

## Route: `/sessions/:id/review`

### 목적
- 질문 적합성, 근거성, 가독성을 검토하고 수정 우선순위를 제시

### 진입 조건
- 세션 존재
- 수동 혹은 자동 검토 결과 1개 이상 존재

### 이탈 조건
- `/sessions/:id`로 복귀

### 정보 구조
- 상단 review summary
- 즉시 확인 필요 섹션
- 우선 수정 항목 섹션
- 문항별 리뷰 아코디언
- unresolved `INFERRED` 섹션
- ready-to-submit 섹션

### 핵심 CTA
- `에디터에서 수정`
- `최종본 복사`

### 필드 목록
- 리뷰 필터
  - 전체 / 질문 적합성 / 근거성 / 가독성

### 검증 규칙
- unresolved `INFERRED`가 있으면 `최종본 복사` disabled
- 문항 미작성 상태가 있으면 제출 준비 완료 배지 미노출

### 비정상 상태
- 검토 결과 없음
- 검토 결과 로드 실패

### 단축키
- `E`: 에디터로 돌아가기
- `C`: 최종본 복사

### 모바일 규칙
- summary card를 세로 스택
- 문항별 리뷰는 full-width accordion

## Route: `/library`

### 목적
- 자기소개 작성에 쓰는 자료를 업로드, 분류, 검색, 미리보기

### 진입 조건
- 로그인 완료

### 이탈 조건
- 자료 선택 후 새 세션 생성
- 기존 세션으로 자료 연결

### 정보 구조
- 상단 page header
- 업로드 dropzone/card
- 필터/검색 바
- 자료 리스트
- 우측 preview pane

### 핵심 CTA
- `파일 업로드`
- `선택 자료로 새 세션 시작`

### 필드 목록
- 파일 업로드
  - accept: `.pdf,.docx,.md,.txt`
- 자료 유형
  - 이력서 / 포트폴리오 / 기존 자소서 / 메모
- 제목
  - default: 파일명
- 메모
  - optional
- 검색어
- 유형 필터

### 검증 규칙
- 스캔 PDF는 unsupported error
- 유형은 업로드 완료 전 선택 필수
- 파일 본문 추출 실패 시 업로드는 완료하되, preview unavailable badge 표시

### 비정상 상태
- 업로드 실패
- 추출 실패
- 미리보기 로드 실패

### 단축키
- `U`: 업로드 dialog 열기
- `/`: 검색 포커스

### 모바일 규칙
- preview pane을 하단 sheet로 이동
- 업로드 카드는 full-width

## Route: `/settings`

### 목적
- 계정 정보, 기본 작성 환경, 데이터 관리 정책 설정

### 진입 조건
- 로그인 완료

### 이탈 조건
- 저장 후 현재 화면 유지
- 로그아웃 시 `/login`

### 정보 구조
- 계정 섹션
- 작성 기본값 섹션
- 데이터 관리 섹션
- 시스템 정보 섹션

### 핵심 CTA
- `설정 저장`
- `전체 데이터 삭제`
- `로그아웃`

### 필드 목록
- 이메일
- 현재 비밀번호
- 새 비밀번호
- 기본 글자 수 기준
  - 공백 포함 / 공백 제외
- 기본 자동 검토
  - ON / OFF
- 기본 자동 반영
  - ON / OFF
- 기본 패널 레이아웃
  - SPLIT / FOCUS

### 검증 규칙
- 비밀번호 변경 시 현재 비밀번호 입력 필수
- 새 비밀번호는 8자 이상
- 기본 자동 반영 ON 설정 시 경고 확인 필요

### 비정상 상태
- 저장 실패
- 계정 정보 로드 실패
- 전체 삭제 실패

### 단축키
- `Cmd/Ctrl+S`: 설정 저장

### 모바일 규칙
- 섹션을 아코디언으로 축약
- destructive action은 full-screen confirm sheet 사용
