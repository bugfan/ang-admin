import urllib.request
import json
import os

api_key = os.environ.get("GEMINI_API_KEY", "")
if not api_key:
    print("NO KEY")
    exit(1)

url = f"https://generativelanguage.googleapis.com/v1beta/models?key={api_key}"
req = urllib.request.Request(url)
with urllib.request.urlopen(req) as response:
    data = json.loads(response.read())
    for m in data.get("models", []):
        if "embedContent" in m.get("supportedGenerationMethods", []):
            print(m["name"])
