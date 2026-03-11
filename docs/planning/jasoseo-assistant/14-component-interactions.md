# 14. Component Interactions

## 목적
이 문서는 핵심 인터랙티브 컴포넌트의 구조, 트리거, 상태 전이, 실패 처리, 반응형 동작을 고정한다.

## 추가 상태 열거형
- `ClaimTagStatus = SUPPORTED | INFERRED | RESOLVED`
- `ApplyMode = MANUAL_REVIEW | AUTO_APPLY`
- `ReviewMode = MANUAL | AUTO | BOTH`
- `PanelLayout = FOCUS | SPLIT | COMPARE`
- `ThemeMode = LIGHT_ONLY`

### PanelLayout 해석
- `SPLIT`: 기본 3패널 작업 상태
- `FOCUS`: side panel 최소화 상태
- `COMPARE`: diff 비교 route에서만 쓰는 시스템 상태

## Global Interaction Rules

### 저장 규칙
- 자동 저장은 아래 세 시점에만 실행된다.
  - 마지막 입력 후 `2초 idle`
  - editor `blur`
  - 페이지 이탈 직전
- 자동 저장은 최신 draft를 업데이트하지만, 매번 version history entry를 생성하지 않는다.
- version history entry는 아래 이벤트에서만 생성한다.
  - 사용자가 명시적으로 `버전 저장`
  - diff 적용 직전
  - 자동 반영 일괄 적용 직전
  - version restore 직전
  - 최종 확정 직전

### 검토 규칙
- auto review ON일 때만 동작한다.
- 자동 검토는 아래 조건을 모두 만족할 때 실행된다.
  - 마지막 저장이 성공 상태
  - 마지막 본문 변경 이후 `3초 idle`
  - 한글 IME 조합 상태가 아님
  - 현재 diff 비교 화면이 아님
- 검토 중 추가 입력이 발생하면 현재 검토는 유지하고, 종료 후 1회 재실행한다.

### IME 규칙
- `compositionstart` 동안 autosave와 autoreview 모두 억제한다.
- `compositionend` 직후 dirty 상태면 autosave 타이머를 다시 시작한다.

### Undo / Snapshot 규칙
- 모든 apply action은 undo 가능한 snapshot을 남긴다.
- auto apply ON일 때도 동일하다.
- toast의 `되돌리기`는 가장 최근 mutation snapshot만 복원한다.
- version history의 `복원`은 현재 draft를 즉시 덮어쓰지 않고, 먼저 confirm dialog를 띄운다.

### Focus Management
- dialog 닫힘 후 focus는 trigger control로 복귀한다.
- diff 적용 후 focus는 첫 번째 변경 문단으로 이동한다.
- review card에서 `에디터에서 수정` 클릭 시 대응 문단으로 스크롤 + focus 이동한다.

## Editor Workspace

### 구조
- context bar
- question tabs
- source panel
- main editor canvas
- right utility panel
- sticky action bar

### 상태
- `idle`
- `dirty`
- `saving`
- `save_error`
- `reviewing`
- `blocked_by_inferred`

### 트리거
- typing
- paste
- paragraph selection
- tab change
- panel toggle

### 상태 전이
- `idle -> dirty`: 본문 변경
- `dirty -> saving`: 2초 idle 또는 blur
- `saving -> idle`: 저장 성공
- `saving -> save_error`: 저장 실패
- `idle -> reviewing`: 수동/자동 검토 시작
- `reviewing -> idle`: 검토 성공

### 실패 처리
- 저장 실패 시 상단 inline banner + `다시 저장`
- 세션 로드 실패 시 full-page error state

### 반응형
- `>=1280`: 3패널
- `1024~1279`: AI 패널 폭 축소
- `<1024`: AI 패널 sheet
- `<768`: source panel도 sheet

## Question Tabs

### 구조
- 질문 번호
- 축약 제목
- 글자 수 상태
- issue dot

### 상태
- `default`
- `active`
- `warning`
- `complete`

### 트리거
- click
- `Alt+1..9`
- mobile horizontal swipe

### 규칙
- 탭 라벨은 20자 이내로 절단
- 글자 수 초과 시 amber badge
- unresolved `INFERRED`가 있는 문항은 issue dot 우선 표시

## ClaimTag

### 구조
- inline underline
- gutter badge
- inspector sheet / popover

### 상태별 시각 규칙
- `SUPPORTED`
  - 기본 읽기 상태에서는 강조 없음
  - hover 또는 inspector 모드에서만 자료 연결 highlight
- `INFERRED`
  - amber dotted underline
  - `AI 추론 추가` badge
- `RESOLVED`
  - 기본 읽기 상태 강조 없음
  - history/inspection에서만 `확정됨` 메타 표시

### 트리거
- click underline
- click badge
- keyboard focus via inspector list

### 액션
- `사실 확인`
- `문장 수정`
- `삭제`

### 실패 처리
- linked source 찾기 실패 시 `근거 정보를 불러오지 못했습니다.` 표시

### 반응형
- desktop: popover
- mobile: bottom sheet

## AI Chat Panel

### 구조
- tool shortcuts row
- conversation list
- suggestion cards
- composer

### 메시지 타입
- user
- assistant
- action_result
- system_hint

### 입력 규칙
- `Enter`: 전송
- `Shift+Enter`: 줄바꿈
- scope selector와 현재 문항 컨텍스트를 항상 함께 보낸다.

### 응답 규칙
- assistant가 본문 변경을 제안하면 suggestion card + compare action을 만든다.
- assistant가 자료 부족을 감지하면 질문 card를 우선 출력한다.
- assistant는 수동 적용 모드에서 범위 밖 변경을 본문에 직접 반영하지 않는다.

### 실패 처리
- AI 응답 실패 시 `다시 시도` 버튼이 있는 inline error bubble 표시

### 반응형
- desktop: 우측 고정 패널
- mobile: full-height sheet

## Tool Menu

### 기본 도구
- `다른 소재 찾기`
- `질문 적합성 개선`
- `가독성 개선`
- `분량 맞추기`
- `톤 조정`

### 동작 원리
- 도구 클릭 즉시 hidden prompt template + 현재 scope + 현재 question context를 composer에 주입한다.
- 사용자 검토 후 바로 전송하거나, 수정 후 전송한다.

### 프롬프트 템플릿
- 다른 소재 찾기:
  - `현재 문항과 선택 범위를 유지한 채, 자료에 있는 경험 중 대체 가능한 소재 3개를 제안하세요. 자료에 없는 사실은 쓰지 말고, 없으면 사용자에게 질문하세요.`
- 질문 적합성 개선:
  - `선택 범위만 수정하고, 현재 문항이 요구하는 역량에 더 직접 답하도록 재작성하세요. 범위 밖 문장은 건드리지 마세요.`
- 가독성 개선:
  - `선택 범위만 유지한 채 문장을 더 짧고 명확하게 다듬으세요. 의미가 바뀌면 안 됩니다.`
- 분량 맞추기:
  - `현재 문항의 글자 수 제한 안에 들어오도록 선택 범위를 압축하거나 확장하세요. 핵심 근거는 유지하세요.`
- 톤 조정:
  - `과장 없이 자신감 있는 톤으로 선택 범위를 다듬으세요. 추상 표현보다 검증 가능한 표현을 우선하세요.`

### 반응형
- desktop: tool pills
- mobile: horizontal scroll chips

## Upload Card

### 구조
- upload trigger
- accepted format hint
- metadata form
- extraction status

### 상태
- `idle`
- `uploading`
- `extracting`
- `ready`
- `error`

### 검증
- accept: `.pdf,.docx,.md,.txt`
- scanned PDF는 error
- asset type 선택 전 완료 불가

### 실패 처리
- 업로드 실패: `업로드에 실패했습니다. 다시 시도해 주세요.`
- 추출 실패: 업로드 완료 + preview unavailable badge

## Diff Viewer

### 구조
- metadata header
- original pane
- suggested pane
- chunk navigator
- sticky action footer

### 상태
- `ready`
- `applying`
- `apply_error`
- `empty`

### 규칙
- 삭제 내용은 red strike
- 추가 내용은 green underline
- 유지 내용은 기본 스타일
- 현재 chunk는 양쪽 pane에서 동일 위치 강조

### 트리거
- apply
- reject
- next / prev chunk

### 실패 처리
- apply 실패 시 footer 위 error banner + retry

### 반응형
- desktop: 2 columns
- mobile: stacked accordion

## QuickReviewDrawer

### 구조
- 점수 summary
- top 3 issues
- `전체 리뷰 보기` CTA

### 규칙
- editor 안에서는 top 3만 노출
- 문항별 전체 분석은 `/sessions/:id/review`에서만 제공

## Review Card

### 구조
- severity badge
- issue title
- why
- fix action
- jump target

### severity
- blocking
- important
- polish

### 트리거
- `에디터에서 수정`
- `관련 문장 보기`

### 실패 처리
- target paragraph 찾기 실패 시 해당 question 탭까지만 이동

## Version History

### 구조
- timestamp
- version label
- source
  - manual save
  - before apply
  - before auto apply
  - before restore
  - final confirm

### 규칙
- auto save 자체는 version list를 오염시키지 않는다.
- restore는 항상 confirm dialog를 거친다.
- restore 후에는 `복원된 버전`이라는 새 history entry를 남긴다.

### 반응형
- desktop: right drawer
- mobile: full-screen sheet

## Toast

### 타입
- success
- info
- warning
- error

### 규칙
- auto dismiss `4초`
- error toast는 auto dismiss하지 않는다.
- undo가 있는 toast는 버튼 포함
- `aria-live="polite"`

## Confirm Dialog

### 사용 시점
- 전체 데이터 삭제
- 세션 삭제
- version restore
- 미저장 상태에서 이탈
- 자동 반영 기본값 ON 저장

### 구조
- title
- impact summary
- primary action
- secondary action

### 규칙
- destructive primary는 danger style
- default focus는 safe action
