# HTMX 2.0.0 Migration & Implementation Guide

This document serves as the **AUTHORITATIVE SOURCE** for all HTMX implementations in this project. All agents must strictly adhere to HTMX 2.0.0 standards and avoid deprecated 1.x patterns.

## 1. Critical Breaking Changes (1.x -> 2.0)

### 1.1 Extensions Moved to Separate Repo
*   **Change:** Core features like **Server Sent Events (SSE)** and **WebSockets (WS)** are **NO LONGER** in the core `htmx.js` file.
*   **Action:** You must explicitly include the extension script if needed.
*   **Example:**
    ```html
    <!-- Core -->
    <script src="https://unpkg.com/htmx.org@2.0.0"></script>
    <!-- Extension (if needed) -->
    <script src="https://unpkg.com/htmx-ext-sse@2.0.0/sse.js"></script>
    ```

### 1.2 HTTP DELETE Requests
*   **Change:** `DELETE` requests now behave like `GET` requests by default. They **do not** send a form body; they send parameters in the URL query string.
*   **Action:** If your backend expects a body for DELETE, you must configure `htmx.config.methodsThatUseUrlParams`.
*   **Default Behavior:** `hx-delete="/resource" hx-vals='{"id": 1}'` -> `DELETE /resource?id=1`

### 1.3 `hx-on` Syntax
*   **Change:** The old `hx-on` attribute (without colons) is deprecated/removed in favor of specific event listeners.
*   **Requirement:** Use `hx-on:event-name`.
*   **Example:**
    *   ❌ `hx-on="click: alert('hi')"`
    *   ✅ `hx-on:click="alert('hi')"`
    *   ✅ `hx-on:htmx:before-request="..."` (Note the kebab-case for HTMX events)

### 1.4 Removed Attributes
*   ❌ `hx-sse` (Use the `sse` extension + `hx-ext="sse"`)
*   ❌ `hx-ws` (Use the `ws` extension + `hx-ext="ws"`)
*   ❌ `hx-vars` (Use `hx-vals`)

## 2. Standard 2.0 Syntax Reference

### 2.1 Basic Requests
```html
<!-- GET -->
<button hx-get="/api/resource" hx-target="#result">Load</button>

<!-- POST with Values -->
<button hx-post="/api/update" 
        hx-vals='{"myVal": "hello"}'
        hx-target="#result">
    Update
</button>
```

### 2.2 Event Handling (`hx-on:`)
Use `hx-on:` followed by the event name.
```html
<button hx-delete="/item/1"
        hx-confirm="Are you sure?"
        hx-on:htmx:after-request="if(event.detail.successful) this.remove()">
    Delete
</button>
```

### 2.3 Swapping
Default is `innerHTML`.
*   `outerHTML`: Replaces the target element itself.
*   `beforeend`: Appends inside the target.
*   `delete`: Deletes the target (useful for removing items).

## 3. Project Configuration
*   **CDN:** `https://unpkg.com/htmx.org@2.0.0`
*   **Extensions:** Only import what is strictly necessary.
