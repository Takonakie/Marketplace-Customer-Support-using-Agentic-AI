import re

def validate_input(text: str) -> str:
    if len(text) > 2000:
        return "Maaf, pesan Anda terlalu panjang. Mohon persingkat pertanyaan Anda."
    
    injection_patterns = [
        r"ignore (all|previous) (instructions|prompts)",
        r"system prompt",
        r"abaikan (semua|pesan|instruksi) (sebelumnya|awal)",
        r"forget (your|all) (rules|instructions)",
        r"you are now (an|a|admin|developer|root)",
        r"mode (jailbreak|developer|dan)",
    ]
    for pattern in injection_patterns:
        if re.search(pattern, text, re.IGNORECASE):
            return "Maaf, permintaan Anda tidak dapat diproses demi alasan keamanan."
    
    return None

def sanitize_output(response_text: str) -> str:
    if len(response_text) > 4000:
        response_text = response_text[:3997] + "..."
    return response_text
