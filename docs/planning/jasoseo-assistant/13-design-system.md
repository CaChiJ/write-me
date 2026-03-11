# 13. Design System

## 디자인 방향

### 스타일 키워드
- Editorial Productivity
- Warm Paper
- Calm Confidence
- Evidence-first Workspace

### 금지 방향
- 네온 계열의 AI 제품 감성
- 보라색 중심 그라데이션
- 과도한 glassmorphism
- MVP 범위를 벗어난 dark mode 선행 설계
- 이모지 아이콘

## 테마 정책
- `ThemeMode = LIGHT_ONLY`
- MVP에서는 light theme만 구현한다.
- dark theme 토큰은 정의하지 않는다.

## 핵심 시각 원칙
1. 문서 작성 도구처럼 차분해야 한다.
2. 정보 위계는 색보다 spacing과 typography로 만든다.
3. 경고는 명확하되 공격적이지 않아야 한다.
4. AI 기능은 장식이 아니라 작업 보조 패널처럼 보여야 한다.
5. 종이 같은 배경과 정돈된 패널 대비로 장시간 사용 피로를 줄인다.

## Color Tokens

| Token | Value | Usage |
|---|---|---|
| `color.page.bg` | `#F6F1E8` | 앱 최외곽 배경 |
| `color.surface.base` | `#FFFDF9` | 기본 카드/패널 배경 |
| `color.surface.subtle` | `#FAF6EF` | rail, 보조 패널, 비활성 영역 |
| `color.surface.raised` | `#FFFFFF` | modal, dialog, active pane |
| `color.text.primary` | `#1F2733` | 본문, 제목 |
| `color.text.secondary` | `#5B6573` | 설명, 메타 정보 |
| `color.text.inverse` | `#FFFDF9` | 진한 배경 위 텍스트 |
| `color.border.subtle` | `#E7DED2` | 기본 경계선 |
| `color.border.strong` | `#D7CCBE` | 강조 경계선 |
| `color.accent.primary` | `#C7792B` | primary CTA, active highlight |
| `color.accent.primary.hover` | `#AE6927` | primary hover |
| `color.accent.secondary` | `#2D7A78` | 선택 상태, secondary CTA, focus ring |
| `color.accent.secondary.soft` | `#E5F3F1` | 보조 배경 강조 |
| `color.warning.base` | `#D97706` | `INFERRED`, 글자수 초과, 주의 |
| `color.warning.soft` | `#FFF1DD` | warning badge background |
| `color.danger.base` | `#B42318` | 삭제, 치명적 오류 |
| `color.danger.soft` | `#FEE4E2` | danger background |
| `color.success.base` | `#18794E` | 저장 성공, 완료 상태 |
| `color.success.soft` | `#E7F7EE` | success background |
| `color.info.base` | `#3659A2` | 일반 정보 배너 |
| `color.info.soft` | `#EAF0FF` | info background |

## Typography

### Font Family
- UI / Body: `SUIT Variable`, fallback `Pretendard Variable`, `-apple-system`, `sans-serif`
- Section / Display: `MaruBuri`, fallback `serif`
- Numeric / tabular: `SUIT Variable` with tabular numerals enabled

### Type Scale

| Token | Size / Line Height | Weight | Usage |
|---|---|---|---|
| `type.display.lg` | `40 / 52` | 700 | 로그인 브랜드, 대형 섹션 헤더 |
| `type.display.md` | `32 / 44` | 700 | 대시보드 헤더 |
| `type.heading.lg` | `28 / 38` | 700 | 페이지 제목 |
| `type.heading.md` | `24 / 34` | 700 | 패널 제목 |
| `type.heading.sm` | `20 / 30` | 600 | 카드 제목 |
| `type.body.lg` | `18 / 30` | 400 | 중요한 본문 |
| `type.body.md` | `16 / 26` | 400 | 기본 본문 |
| `type.body.sm` | `14 / 22` | 400 | 설명 텍스트 |
| `type.label.lg` | `16 / 24` | 600 | 큰 버튼, 탭 |
| `type.label.md` | `14 / 20` | 600 | 일반 버튼, 필드 라벨 |
| `type.label.sm` | `12 / 18` | 600 | badge, helper label |

### Typography Rules
- body text minimum size: `16px`
- paragraph max line length on desktop: `65~75 chars`
- paragraph max line length on mobile: `35~60 chars`
- body line-height minimum: `1.5`
- placeholder-only label 금지

## Spacing and Radius

### Spacing Scale
- base unit: `8px`
- token set: `4, 8, 12, 16, 24, 32, 40, 48, 64`

### Radius

| Token | Value | Usage |
|---|---|---|
| `radius.sm` | `12px` | input, small card, badge |
| `radius.md` | `16px` | panel, modal body, upload card |
| `radius.lg` | `24px` | hero card, large surface |

## Elevation

| Token | Value | Usage |
|---|---|---|
| `shadow.sm` | `0 1px 2px rgba(31,39,51,0.06)` | 기본 카드 |
| `shadow.md` | `0 8px 24px rgba(31,39,51,0.10)` | sticky bar, raised panel |
| `shadow.lg` | `0 20px 40px rgba(31,39,51,0.12)` | dialog, compare overlay |

## Motion

### Motion Tokens
- `motion.fast = 160ms ease-out`
- `motion.base = 240ms ease-out`

### Motion Rules
- transform + opacity만 animate
- width / height / top / left animation 금지
- press feedback는 100ms 이내에 시작
- reduced motion 활성화 시 모든 non-essential motion 제거
- animation은 decorative가 아니라 상태 전이에만 사용

## Layout Tokens

| Token | Value |
|---|---|
| `bp.wide` | `1440px` |
| `bp.desktop` | `1280px` |
| `bp.compact` | `1024px` |
| `bp.tablet` | `768px` |
| `bp.mobile` | `390px` |
| `layout.navRail` | `88px` |
| `layout.sourcePanel` | `280px` |
| `layout.editorMin` | `640px` |
| `layout.aiPanel` | `360px` |
| `layout.pageGutter.desktop` | `32px` |
| `layout.pageGutter.tablet` | `24px` |
| `layout.pageGutter.mobile` | `16px` |

## Iconography
- icon set: `Lucide`
- stroke width: `1.75`
- 기본 크기:
  - `16px`: inline helper
  - `20px`: button leading icon
  - `24px`: nav, toolbar
- emoji를 구조적 아이콘으로 사용하지 않는다.
- icon-only button은 `aria-label` 필수

## Component Visual Rules

### Buttons
- Primary:
  - bg `color.accent.primary`
  - text `color.text.inverse`
  - hover `color.accent.primary.hover`
- Secondary:
  - bg `color.surface.base`
  - border `color.border.strong`
  - text `color.text.primary`
- Danger:
  - bg `color.danger.base`
  - text `color.text.inverse`
- height:
  - large `48px`
  - default `44px`
  - small `36px`

### Inputs
- background `color.surface.raised`
- border `1px solid color.border.subtle`
- focus ring `2px color.accent.secondary`
- error border `color.danger.base`
- helper text는 항상 field 바로 아래

### Cards and Panels
- 카드 배경은 `surface.base`
- 우측 AI 패널과 좌측 자료 패널은 `surface.subtle`
- sticky bars는 `surface.raised + shadow.md`
- panel 간 간격은 최소 `16px`

### Tabs
- inactive: text `color.text.secondary`
- active: text `color.text.primary`
- active underline: `2px color.accent.primary`
- badge warning이 있을 경우 active underline과 겹치지 않게 우측 badge 분리

### ClaimTag Styles
- `SUPPORTED`: 기본 본문 스타일, hover 시만 자료 연결 highlight
- `INFERRED`: `2px dotted color.warning.base` underline + soft warning badge
- `RESOLVED`: 기본 읽기 상태에서는 강조 제거, history/inspection 모드에서만 subtle indicator

## Accessibility Baseline
- 텍스트 contrast minimum `4.5:1`
- large text contrast minimum `3:1`
- touch target minimum `44x44`
- visible focus ring 항상 유지
- color-only meaning 금지
- keyboard 접근 가능한 tab order 유지
- destructive CTA는 시각적으로 분리
- toast는 focus를 뺏지 않고 `aria-live="polite"`

## 반응형 원칙
- `>=1280`: 3패널 fully visible
- `1024~1279`: 3패널 유지, 내부 padding 축소
- `768~1023`: 우측 AI 패널 drawer화
- `<768`: 좌측/우측 패널 모두 sheet화, 본문 단일 열
- mobile에서도 body text 16px 유지
