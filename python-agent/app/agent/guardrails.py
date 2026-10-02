import re

def validate_input(text: str) -> str:
    if len(text) > 2000:
        return "Maaf, pesan Anda terlalu panjang. Mohon persingkat pertanyaan Anda."
    
    injection_patterns = [
        r"ignore previous instructions",
        r"system prompt",
        r"abai pesan sebelumnya"
    ]
    for pattern in injection_patterns:
        if re.search(pattern, text, re.IGNORECASE):
            return "Maaf, permintaan Anda tidak dapat diproses."
    
    return None

def sanitize_output(response_text: str) -> str:
    if len(response_text) > 4000:
        response_text = response_text[:3997] + "..."
    return response_text
