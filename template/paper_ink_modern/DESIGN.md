# Design System Document: The Ethereal Archive

## 1. Overview & Creative North Star: "The Digital Curator"
This design system moves away from the cluttered, ad-heavy tropes of digital media libraries. Our North Star is **"The Digital Curator"**—an editorial-first approach that treats manga not just as content, but as art. 

We break the "standard tech" template through **expansive white space, intentional asymmetry, and tonal depth.** Instead of rigid grids and harsh borders, we use a "Soft Minimalism" philosophy. The interface should feel like a premium gallery space: silent, sophisticated, and secondary to the vibrant artwork it hosts. By utilizing high-contrast typography scales and overlapping "glass" layers, we create a sense of tactile luxury.

---

## 2. Colors: Tonal Architecture
We move beyond flat hex codes to an architecture of light. Colors are used to guide the eye and define hierarchy without visual noise.

### The "No-Line" Rule
**Explicit Instruction:** 1px solid borders are prohibited for sectioning. Structural boundaries must be defined solely through background color shifts. Use `surface-container-low` (#f3f4f5) for large section backgrounds and `surface` (#f8f9fa) for the primary canvas.

### Surface Hierarchy & Nesting
Treat the UI as a series of physical layers. Use the tiers to create natural depth:
- **Base Layer:** `surface` (#f8f9fa)
- **Secondary Sections:** `surface-container-low` (#f3f4f5)
- **Interactive Cards:** `surface-container-lowest` (#ffffff)
- **Floating Overlays:** `surface-bright` (#f8f9fa) with Glassmorphism.

### Signature Textures & Gradients
To provide "visual soul," use a signature linear gradient (45-degree) transitioning from `primary` (#24389c) to `primary-container` (#3f51b5). This is reserved exclusively for:
- Main Call-to-Action (CTA) backgrounds.
- Hero section accents.
- Progress indicators.

### Glassmorphism & Depth
For floating navigation bars or filter overlays, use `surface_container_lowest` at 80% opacity with a `24px` backdrop-blur. This softens the interface and prevents the "pasted-on" look of traditional drawers.

---

### 3. Typography: Editorial Authority
The type system pairs the geometric precision of **Manrope** for display with the high legibility of **Inter** for utility.

*   **Display (Manrope):** Large, bold, and authoritative. Use `display-lg` (3.5rem) with tight letter-spacing (-0.02em) for hero manga titles to create an editorial feel.
*   **Headlines (Manrope):** `headline-md` (1.75rem) should be used for section titles, often placed with asymmetrical padding to break the grid.
*   **Body (Inter):** All body text uses `body-md` (0.875rem) for a modern, tech-forward aesthetic. Ensure a line height of 1.6 for maximum breathability.
*   **Labels (Inter):** Use `label-md` (0.75rem) in `secondary` (#4a626d) for metadata (e.g., Chapter counts, Genres).

---

## 4. Elevation & Depth: Tonal Layering
We do not use shadows to create "pop"; we use them to simulate "presence."

*   **The Layering Principle:** Place a `surface-container-lowest` (#ffffff) card on a `surface-container-low` (#f3f4f5) background. This creates a "soft lift" that is felt rather than seen.
*   **Ambient Shadows:** When a card requires a floating state (e.g., on hover), use a shadow: `0 20px 40px rgba(25, 28, 29, 0.04)`. The shadow must be tinted with the `on-surface` color to look natural.
*   **The "Ghost Border" Fallback:** If accessibility requires a border, use `outline-variant` (#c5c5d4) at 15% opacity. Never use 100% opaque lines.
*   **Asymmetric Cards:** Manga covers should "bleed" out of their containers or sit slightly offset (e.g., -12px margin) to break the boxy feel of traditional web layouts.

---

## 5. Components: Modern Primitives

### Buttons
*   **Primary:** Gradient (Primary to Primary-Container), white text, `lg` (0.5rem) roundedness. No shadow.
*   **Secondary:** `surface-container-high` (#e7e8e9) background, `primary` text.
*   **Tertiary:** Transparent background, `primary` text, no border.

### Manga Cards & Lists
*   **Constraint:** **Forbidden use of divider lines.**
*   **Layout:** Use `xl` (0.75rem) corner radius for manga covers.
*   **Separation:** Separate list items using `1.5rem` of vertical white space or a subtle shift to `surface-container-low` on hover.

### Input Fields
*   **Style:** Minimalist. No bottom line or full border. Use `surface-container-highest` (#e1e3e4) as a solid background fill with `md` (0.375rem) rounding.
*   **Focus:** A subtle `outline` (#757684) at 20% opacity.

### Navigation (The Floating Bar)
*   Instead of a top-fixed header, use a centered, floating navigation "pill" using the `full` (9999px) roundedness scale, `surface-container-lowest` background at 90% opacity, and a `glassmorphism` blur.

---

## 6. Do’s and Don’ts

### Do:
*   **Do** use asymmetrical margins (e.g., more padding on the left than the right) for title sections to mimic high-end magazine layouts.
*   **Do** allow manga cover art to dictate the "vibe" of the page by using semi-transparent containers that let art colors bleed through.
*   **Do** use `primary-fixed` (#dee0ff) for subtle "New" or "Hot" badges.

### Don't:
*   **Don't** use black (#000000). Use `on-surface` (#191c1d) for all "black" text to maintain the soft aesthetic.
*   **Don't** use standard "Drop Shadows." Only use the Ambient Shadow spec provided.
*   **Don't** cram content. If a section feels busy, double the padding. This system relies on "The Breath"—the space between elements.
*   **Don't** use icons with heavy fills. Use thin-stroke (1.5px) outline icons to match the Inter/Manrope weight.