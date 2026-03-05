import tkinter as tk
from tkinter import ttk, scrolledtext
import requests
import threading
import json

AI_CORE_URL = "http://localhost:8080/api/command"

class OffTopicDashboard:
    def __init__(self, root):
        self.root = root
        self.root.title("OffTopic Linux Dashboard")
        self.root.geometry("800x600")
        self.root.configure(bg="#1e1e2e")
        self.root.option_add("*Font", "Consolas 10")
        
        # Style
        style = ttk.Style()
        style.theme_use("clam")
        style.configure("TButton", padding=10, font=("Consolas", 11, "bold"), background="#313244", foreground="#cdd6f4")
        style.map("TButton", background=[("active", "#45475a")])
        style.configure("TLabel", background="#1e1e2e", foreground="#cdd6f4", font=("Consolas", 14, "bold"))
        style.configure("TFrame", background="#1e1e2e")
        
        # Title Label
        title = ttk.Label(self.root, text="[ OffTopic Linux // AI-Core Control Panel ]")
        title.pack(pady=20)
        
        # Buttons Frame
        btn_frame = ttk.Frame(self.root)
        btn_frame.pack(pady=10)
        
        # 1-Click Buttons
        self.btn_recon = ttk.Button(btn_frame, text="1-Click Recon", command=lambda: self.run_task("recon"))
        self.btn_recon.grid(row=0, column=0, padx=10)
        
        self.btn_vuln = ttk.Button(btn_frame, text="1-Click Vuln Match", command=lambda: self.run_task("vuln_match"))
        self.btn_vuln.grid(row=0, column=1, padx=10)
        
        self.btn_stealth = ttk.Button(btn_frame, text="1-Click Stealth Mode", command=lambda: self.run_task("stealth_mode"))
        self.btn_stealth.grid(row=0, column=2, padx=10)
        
        # Log Output Frame
        log_frame = ttk.Frame(self.root)
        log_frame.pack(fill=tk.BOTH, expand=True, padx=20, pady=20)
        
        log_label = ttk.Label(log_frame, text="System Logs:", font=("Consolas", 11))
        log_label.pack(anchor="w")
        
        self.log_area = scrolledtext.ScrolledText(log_frame, bg="#11111b", fg="#a6e3a1", font=("Consolas", 10), wrap=tk.WORD)
        self.log_area.pack(fill=tk.BOTH, expand=True, pady=5)
        
        self.log_message("[*] Dashboard Initialized. Connecting to AI-Core at localhost:8080...")

    def log_message(self, message):
        self.log_area.insert(tk.END, message + "\n")
        self.log_area.see(tk.END)

    def run_task(self, task_type):
        self.log_message(f"[*] Dispatching task: {task_type.upper()}...")
        # Run in thread to prevent UI freezing
        threading.Thread(target=self._send_request, args=(task_type,), daemon=True).start()

    def _send_request(self, task_type):
        try:
            payload = {"action": task_type, "timestamp": "now"}
            response = requests.post(AI_CORE_URL, json=payload, timeout=15)
            
            if response.status_code == 200:
                self.log_message(f"[+] SUCCESS ({task_type}):\n{json.dumps(response.json(), indent=2)}")
            else:
                self.log_message(f"[-] HTTP {response.status_code} Error: {response.text}")
                
        except requests.exceptions.ConnectionError:
            self.log_message(f"[!] CONNECTION ERROR: AI-Core is not running on {AI_CORE_URL}. Please start the backend.")
        except Exception as e:
            self.log_message(f"[!] UNEXPECTED ERROR: {str(e)}")

if __name__ == "__main__":
    app_root = tk.Tk()
    dashboard = OffTopicDashboard(app_root)
    app_root.mainloop()
