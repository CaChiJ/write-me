# 00. Agent Handoff (Start Here)

## 목적
이 문서는 **컨텍스트 없이 투입된 다른 에이전트**가 오해 없이 작업을 이어가기 위한 시작점이다.

## 현재 기준 버전
- 기획 문서 버전: `v0.4`
- 기준 일자: `2026-03-11`

## 단일 진실 문서(SSOT)
1. 확정 결정사항: [09-decisions-v0.3.1.md](./09-decisions-v0.3.1.md)
2. 화면 명세: [11-screen-specs.md](./11-screen-specs.md)
3. 디자인 시스템: [13-design-system.md](./13-design-system.md)
4. 상호작용 규칙: [14-component-interactions.md](./14-component-interactions.md)
5. 와이어프레임: [12-wireframes.md](./12-wireframes.md)
6. 카피/상태 매트릭스: [15-copy-and-state-matrix.md](./15-copy-and-state-matrix.md)

## 지원 문서
1. 사용자 서사 기준: [10-storyboard.md](./10-storyboard.md)
2. 제품 방향/범위: [01-product-scope.md](./01-product-scope.md)

## 충돌 해석 규칙
1. 문서 간 충돌 시 `09-decisions-v0.3.1.md`가 최우선
2. 그다음 `11-screen-specs.md`, `14-component-interactions.md`, `13-design-system.md` 순으로 해석
3. 확정사항 변경이 필요하면 먼저 사용자 확인을 받는다

## 절대 바꾸면 안 되는 확정사항
1. 개인 사용자 1명 기준
2. 1세션 = 1지원 공고
3. 지원 페이지 입력은 텍스트 붙여넣기만
4. 수동 업로드만 지원, 포맷은 PDF/DOCX/MD/TXT
5. 업로드 용량 제한 없음
6. 스캔 PDF(OCR) 비지원
7. 자동 검토 + 수동 검토 동시 지원, 자동 검토 유휴 시간 3초
8. 기본은 제안 후 사용자 적용(좌/우 diff 비교)
9. 자동 반영 ON 시 범위 밖 변경 허용
10. `INFERRED` 문장은 문장별 확인 필수
11. 보관은 영구, 삭제는 즉시 삭제
12. 최소 로그인(단일 관리자 계정)
13. light theme only
14. 아이콘은 SVG 계열만 사용, emoji 아이콘 금지

## 자주 오해되는 포인트
1. 형식별(대기업/스타트업/요약형) 메뉴 분기를 늘리지 않는다.
2. 자동 반영 OFF와 ON의 정책이 다르다.
3. `INFERRED` 허용은 "허위 사실 허용"이 아니라 "표시 + 사용자 확인" 전제다.
4. 내보내기는 복사/텍스트만이 MVP 범위다.
5. 검토 화면은 full page route이고, editor 안에는 quick review drawer가 별도로 있다.
6. MVP에는 `/sessions` 목록 화면이나 `/exports` 전용 화면이 없다. 최근 작업은 `/dashboard`, 내보내기는 세션/검토 화면 액션으로 처리한다.

## 다음 작업 우선순위
1. 문서 기준으로 실제 화면 구현 시작
2. editor tech 선택과 inline annotation 구현 설계
3. session/review/diff 상태 모델을 UI에 연결
4. 한글 IME + autosave/autoreview QA 시나리오 검증

## 작업 전 체크리스트
1. 최신 확정사항 문서(`09-decisions-v0.3.1.md`)를 먼저 확인했는가
2. `11`, `13`, `14`, `15` 문서의 라벨/토큰/상태 규칙을 먼저 확인했는가
3. 제안하는 변경이 확정사항과 충돌하지 않는가
4. 충돌 시 문서 수정보다 사용자 확인이 우선인가
