# 15. Copy And State Matrix

## 목적
이 문서는 Korean-first UI 라벨과 상태별 메시지를 고정한다.  
개발자는 이 문서를 기준으로 버튼명, 오류 문구, 빈 상태 문구, CTA 문구를 그대로 구현한다.

## Global Labels

| 영역 | Label |
|---|---|
| 서비스명 | `Write Me` |
| 저장 성공 | `저장되었습니다.` |
| 저장 실패 | `저장에 실패했습니다. 다시 시도해 주세요.` |
| 다시 시도 | `다시 시도` |
| 취소 | `취소` |
| 적용 | `적용` |
| 적용 안 함 | `적용 안 함` |
| 삭제 | `삭제` |
| 되돌리기 | `되돌리기` |
| 닫기 | `닫기` |

## Navigation Copy

| 메뉴 | Label |
|---|---|
| Dashboard | `대시보드` |
| Library | `자료함` |
| Settings | `설정` |

## Screen Copy

### `/login`

| Element | Copy |
|---|---|
| page title | `로그인` |
| email label | `이메일` |
| password label | `비밀번호` |
| submit | `로그인` |
| helper | `단일 관리자 계정으로 접근합니다.` |

### `/dashboard`

| Element | Copy |
|---|---|
| page title | `대시보드` |
| section title | `자기소개 작성 시작` |
| cta 1 | `지원 페이지 붙여넣고 시작` |
| cta 2 | `기존 초안 붙여넣고 시작` |
| cta 3 | `자료 선택 후 시작` |
| recent title | `최근 작업` |

### `/sessions/new`

| Element | Copy |
|---|---|
| page title | `새 작성 시작` |
| step 1 | `시작 방식 선택` |
| step 2 | `입력` |
| step 3 | `자동 구조화 결과 확인` |
| step 4 | `자료 연결` |
| parse cta | `구조화 실행` |
| final cta | `작성 세션 만들기` |

### `/sessions/:id`

| Element | Copy |
|---|---|
| action review | `검토 열기` |
| action compare | `비교 보기` |
| action finalize | `최종 확정` |
| action unresolved | `미확정 문장 확인` |
| toolbar count mode | `공백 포함` / `공백 제외` |
| toolbar auto review | `자동 검토` |
| toolbar auto apply | `자동 반영` |
| toolbar scope | `선택 문단` / `전체 문서` |

### `/sessions/:id/compare`

| Element | Copy |
|---|---|
| page title | `수정안 비교` |
| apply | `이 수정안 적용` |
| reject | `적용 안 함` |
| prev | `이전` |
| next | `다음` |

### `/sessions/:id/review`

| Element | Copy |
|---|---|
| page title | `리뷰 리포트` |
| blocking title | `즉시 확인 필요` |
| back editor | `에디터에서 수정` |
| copy final | `최종본 복사` |

### `/library`

| Element | Copy |
|---|---|
| page title | `자료 라이브러리` |
| upload | `파일 업로드` |
| search placeholder | `자료 제목 또는 본문 검색` |
| start with selected | `선택 자료로 새 세션 시작` |

### `/settings`

| Element | Copy |
|---|---|
| page title | `설정` |
| save | `설정 저장` |
| delete all | `전체 데이터 삭제` |
| logout | `로그아웃` |

## Validation Copy

| Condition | Copy |
|---|---|
| invalid email | `올바른 이메일 형식을 입력해 주세요.` |
| empty password | `비밀번호를 입력해 주세요.` |
| posting too short | `지원 페이지 텍스트를 조금 더 길게 붙여넣어 주세요.` |
| draft too short | `기존 초안이 너무 짧습니다. 최소 30자 이상 입력해 주세요.` |
| no asset selected | `최소 1개의 자료를 선택해 주세요.` |
| invalid char limit | `글자 수 제한은 1에서 5000 사이의 숫자여야 합니다.` |
| unresolved inferred | `미확정 문장을 모두 확인한 뒤 최종 확정할 수 있습니다.` |
| scanned pdf | `스캔 PDF는 현재 지원하지 않습니다. 텍스트 추출 가능한 파일을 업로드해 주세요.` |
| current password required | `현재 비밀번호를 입력해 주세요.` |
| new password too short | `새 비밀번호는 8자 이상이어야 합니다.` |

## Tool Menu Copy

| Tool | Label | AI Composer Default Copy |
|---|---|---|
| alternative | `다른 소재 찾기` | `이 문항에 더 잘 맞는 다른 소재가 있으면 3개 제안해 줘.` |
| relevance | `질문 적합성 개선` | `선택한 범위만 수정해서 질문 의도에 더 직접 답하게 해 줘.` |
| readability | `가독성 개선` | `선택한 범위만 더 짧고 읽기 쉽게 다듬어 줘.` |
| length | `분량 맞추기` | `이 문항 글자 수 제한 안에 들어오도록 선택한 범위를 조정해 줘.` |
| tone | `톤 조정` | `과장 없이 자신감 있는 톤으로 선택한 범위를 다듬어 줘.` |

## Dialog And Toast Copy

| Situation | Title | Body | Primary |
|---|---|---|---|
| delete all | `전체 데이터를 삭제할까요?` | `세션, 자료, 저장된 버전이 즉시 삭제되며 복구할 수 없습니다.` | `전체 데이터 삭제` |
| unsaved leave | `저장되지 않은 변경이 있습니다.` | `이 페이지를 떠나면 최근 변경이 반영되지 않을 수 있습니다.` | `그래도 나가기` |
| restore version | `이 버전으로 복원할까요?` | `현재 초안은 복원 전 스냅샷으로 저장됩니다.` | `이 버전으로 복원` |
| auto apply enable | `자동 반영을 기본값으로 저장할까요?` | `AI가 비교 단계 없이 본문에 변경을 반영합니다.` | `저장` |

| Toast Type | Copy |
|---|---|
| save success | `저장되었습니다.` |
| version saved | `새 버전이 저장되었습니다.` |
| diff applied | `수정안이 적용되었습니다.` |
| auto applied | `AI 변경이 자동 반영되었습니다.` |
| copied | `최종본이 복사되었습니다.` |
| deleted | `삭제되었습니다.` |

## Screen State Matrix

| Screen | Default | Empty | Loading | Error | Success |
|---|---|---|---|---|---|
| `/login` | 로그인 폼 표시 | 해당 없음 | 버튼 spinner | `로그인에 실패했습니다. 이메일과 비밀번호를 다시 확인해 주세요.` | `/dashboard` 이동 |
| `/dashboard` | 시작 카드 + 최근 작업 | `아직 작성 세션이 없습니다. 첫 세션을 시작해 보세요.` | skeleton cards | `대시보드를 불러오지 못했습니다. 새로고침 후 다시 시도해 주세요.` | 세션 생성 후 success toast |
| `/sessions/new` | 4단계 입력 UI | `연결된 자료가 없습니다. 자료 없이도 계속할 수 있습니다.` | parse spinner | `지원 페이지를 구조화하지 못했습니다. 직접 수정해 주세요.` | 세션 생성 후 editor 이동 |
| `/sessions/:id` | 3패널 편집 UI | `아직 이 문항의 초안이 없습니다.` | saving/reviewing indicator | `세션을 불러오지 못했습니다. 다시 열어 주세요.` | 저장/검토 toast |
| `/sessions/:id/compare` | diff 2컬럼 | `비교할 수정안이 없습니다.` | diff load skeleton | `수정안 비교를 불러오지 못했습니다.` | apply 후 success toast |
| `/sessions/:id/review` | summary + issues | `아직 검토 결과가 없습니다.` | score skeleton | `리뷰 결과를 불러오지 못했습니다.` | 최종본 복사 가능 상태 |
| `/library` | 리스트 + preview | `아직 업로드된 자료가 없습니다.` | list skeleton | `자료 목록을 불러오지 못했습니다.` | 업로드 success toast |
| `/settings` | form sections | 해당 없음 | form skeleton | `설정을 저장하지 못했습니다. 다시 시도해 주세요.` | `설정이 저장되었습니다.` |

## Component State Matrix

| Component | Default | Empty | Loading | Error | Success |
|---|---|---|---|---|---|
| Editor | 본문 표시 | `초안이 없습니다.` | save/review indicator | inline banner | saved indicator |
| QuestionTabs | 라벨 + count | 질문 없음 | tabs shimmer | 해당 없음 | current tab completed dot |
| ClaimTag | 표시 없음 또는 hover link | 해당 없음 | inspector loading | `근거 정보를 불러오지 못했습니다.` | resolved metadata |
| AIChat | message thread | `도구를 선택하거나 직접 요청을 입력해 보세요.` | assistant typing | `응답을 불러오지 못했습니다.` | suggestion card 생성 |
| ToolMenu | 5 tool pills | 해당 없음 | disabled while request active | 해당 없음 | prompt injected |
| UploadCard | idle form | 해당 없음 | upload/extract progress | upload error | asset ready |
| DiffViewer | original/suggested pair | `수정안이 없습니다.` | diff loading | apply error banner | applied toast |
| ReviewCard | issue item | `해당 문항 이슈가 없습니다.` | skeleton row | link target missing warning | jump success |
| VersionHistory | history list | `저장된 버전이 없습니다.` | list loading | `버전 목록을 불러오지 못했습니다.` | restore success toast |
| Toast | auto dismiss | 해당 없음 | 해당 없음 | sticky for error | auto dismiss after 4s |
| ConfirmDialog | title + body + actions | 해당 없음 | primary loading | inline action error | close + toast |

## Error Message Writing Rule
- 모든 에러 문구는 `무엇이 문제인지` + `사용자가 다음에 무엇을 해야 하는지`를 함께 말한다.
- 예시:
  - 나쁨: `오류가 발생했습니다.`
  - 좋음: `자료를 불러오지 못했습니다. 새로고침 후 다시 시도해 주세요.`
