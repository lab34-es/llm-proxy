// LLM Proxy Dashboard - Minimal JS
(function () {
  "use strict";

  // ── Playground ────────────────────────────────────────────────────────
  const form = document.getElementById("pg-form");
  if (!form) return;

  const keyInput = document.getElementById("pg-key");
  const modelInput = document.getElementById("pg-model");
  const systemInput = document.getElementById("pg-system");
  const messagesDiv = document.getElementById("pg-messages");
  const input = document.getElementById("pg-input");
  const sendBtn = document.getElementById("pg-send");
  const clearBtn = document.getElementById("pg-clear");

  let conversation = [];
  let streaming = false;

  function appendMessage(role, text) {
    // Clear placeholder on first message
    if (conversation.length === 0 && messagesDiv.querySelector("p")) {
      messagesDiv.innerHTML = "";
    }

    const div = document.createElement("div");
    div.style.marginBottom = "0.75rem";
    div.style.whiteSpace = "pre-wrap";
    div.style.wordBreak = "break-word";

    const label = document.createElement("strong");
    label.textContent = role === "user" ? "You: " : "Assistant: ";
    label.style.color =
      role === "user"
        ? "var(--pico-primary)"
        : "var(--pico-secondary)";

    const content = document.createElement("span");
    content.textContent = text;

    div.appendChild(label);
    div.appendChild(content);
    messagesDiv.appendChild(div);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;

    return content;
  }

  function clearConversation() {
    conversation = [];
    messagesDiv.innerHTML =
      '<p style="color:var(--pico-muted-color)">Send a message to start a conversation.</p>';
  }

  async function sendMessage() {
    const text = input.value.trim();
    const apiKey = keyInput.value.trim();
    const model = modelInput.value.trim();

    if (!apiKey) {
      alert("Please enter an API key.");
      keyInput.focus();
      return;
    }
    if (!model) {
      alert("Please enter a model name.");
      modelInput.focus();
      return;
    }
    if (!text) return;

    // Add user message
    conversation.push({ role: "user", content: text });
    appendMessage("user", text);
    input.value = "";

    // Build messages array
    const messages = [];
    const sys = systemInput.value.trim();
    if (sys) {
      messages.push({ role: "system", content: sys });
    }
    messages.push(...conversation);

    // Disable input while streaming
    streaming = true;
    sendBtn.disabled = true;
    sendBtn.textContent = "...";
    input.disabled = true;

    const contentSpan = appendMessage("assistant", "");

    try {
      const resp = await fetch("/v1/chat/completions", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + apiKey,
        },
        body: JSON.stringify({
          model: model,
          messages: messages,
          stream: true,
        }),
      });

      if (!resp.ok) {
        const err = await resp.text();
        contentSpan.textContent = "Error: " + err;
        // Remove the failed assistant message from conversation
        return;
      }

      const reader = resp.body.getReader();
      const decoder = new TextDecoder();
      let assistantText = "";
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (!line.startsWith("data: ")) continue;
          const data = line.slice(6).trim();
          if (data === "[DONE]") continue;

          try {
            const parsed = JSON.parse(data);
            const delta = parsed.choices?.[0]?.delta?.content;
            if (delta) {
              assistantText += delta;
              contentSpan.textContent = assistantText;
              messagesDiv.scrollTop = messagesDiv.scrollHeight;
            }
          } catch (e) {
            // skip malformed chunks
          }
        }
      }

      conversation.push({ role: "assistant", content: assistantText });
    } catch (err) {
      contentSpan.textContent = "Error: " + err.message;
    } finally {
      streaming = false;
      sendBtn.disabled = false;
      sendBtn.textContent = "Send";
      input.disabled = false;
      input.focus();
    }
  }

  form.addEventListener("submit", function (e) {
    e.preventDefault();
    if (!streaming) sendMessage();
  });

  input.addEventListener("keydown", function (e) {
    if (e.key === "Enter" && !e.shiftKey && !streaming) {
      e.preventDefault();
      sendMessage();
    }
  });

  if (clearBtn) {
    clearBtn.addEventListener("click", clearConversation);
  }
})();
