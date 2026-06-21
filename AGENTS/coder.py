#!/usr/bin/env python3
"""
Vyzorix Coder - Efficient AI Coding Assistant
=============================================
A lightweight, efficient coding assistant that uses context injection
instead of expensive tool-calling.

Usage:
    export SILICONFLOW_API_KEY="your-key"
    python3 coder.py
    
    Or:
    from coder import Coder
    coder = Coder()
    result = coder.review("Review auth handlers")
"""

import os
import sys
import requests
from pathlib import Path

# Project root
PROJECT_ROOT = Path(__file__).parent.parent

# SiliconFlow config
API_KEY = os.environ.get("SILICONFLOW_API_KEY", "")
API_BASE = "https://api.siliconflow.com/v1"

# Default model
DEFAULT_MODEL = "moonshotai/Kimi-K2.6"

# System prompt
SYSTEM_PROMPT = """You are a senior Go engineer reviewing the Vyzorix Update Server codebase.

Guidelines:
- Be thorough and production-quality
- Report specific issues with file paths and line numbers
- Suggest concrete fixes, not vague recommendations
- Focus on: security, correctness, performance, maintainability

Output format:
- Start with summary
- List issues with severity (HIGH/MEDIUM/LOW)
- End with actionable recommendations"""


class Coder:
    """
    Efficient coding assistant using direct API calls.
    Reads context once, sends ONE request, gets comprehensive response.
    """
    
    def __init__(self, model: str = DEFAULT_MODEL):
        self.model = model
        self.history = []
        
    def review(self, task: str, files: list = None) -> str:
        """
        Review code with context from specified files.
        
        Args:
            task: The review task description
            files: List of file paths to include in context
            
        Returns:
            Review response from the model
        """
        # Build context
        context = self._build_context(files)
        
        # Build prompt
        prompt = f"""## TASK
{task}

## CONTEXT
{context}

## RESPONSE"""
        
        # Make single API call
        response = self._call_model(prompt)
        
        return response
    
    def ask(self, question: str, context_files: list = None) -> str:
        """Ask a question about the codebase."""
        files = context_files or []
        return self.review(question, files)
    
    def fix(self, issue: str, files: list = None) -> str:
        """Get fix suggestions for an issue."""
        files = files or []
        context = self._build_context(files)
        
        prompt = f"""## TASK: Fix Issue
{issue}

## CONTEXT
{context}

## INSTRUCTIONS
1. Identify the exact problem
2. Provide the specific code changes needed
3. Show before/after code snippets

## RESPONSE"""
        
        return self._call_model(prompt)
    
    def refactor(self, task: str, files: list = None) -> str:
        """Get refactoring suggestions."""
        files = files or []
        context = self._build_context(files)
        
        prompt = f"""## TASK: Refactor
{task}

## CONTEXT
{context}

## INSTRUCTIONS
1. Explain the current structure
2. Describe the proposed structure
3. Provide migration steps
4. Show key code changes

## RESPONSE"""
        
        return self._call_model(prompt)
    
    def _build_context(self, files: list = None) -> str:
        """Build context string from files."""
        if not files:
            return "(No files specified)"
        
        context_parts = []
        for file_path in files:
            full_path = PROJECT_ROOT / file_path
            try:
                with open(full_path, 'r') as f:
                    content = f.read()
                # Truncate very long files
                if len(content) > 10000:
                    content = content[:10000] + f"\n\n... [truncated, {len(content)-10000} more chars]"
                context_parts.append(f"### {file_path}\n```go\n{content}\n```")
            except FileNotFoundError:
                context_parts.append(f"### {file_path}\n[FILE NOT FOUND]")
            except Exception as e:
                context_parts.append(f"### {file_path}\n[ERROR: {e}]")
        
        return "\n\n".join(context_parts)
    
    def _call_model(self, prompt: str) -> str:
        """Make a single API call."""
        if not API_KEY:
            return "❌ SILICONFLOW_API_KEY not set!"
        
        headers = {
            "Authorization": f"Bearer {API_KEY}",
            "Content-Type": "application/json"
        }
        
        payload = {
            "model": self.model,
            "messages": [
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": prompt}
            ],
            "temperature": 0.3,
            "max_tokens": 4000
        }
        
        try:
            response = requests.post(
                f"{API_BASE}/chat/completions",
                headers=headers,
                json=payload,
                timeout=120
            )
            
            if response.status_code != 200:
                return f"❌ API Error: {response.status_code} - {response.text}"
            
            data = response.json()
            return data["choices"][0]["message"]["content"]
            
        except requests.exceptions.Timeout:
            return "❌ Request timed out"
        except Exception as e:
            return f"❌ Error: {e}"


# =============================================================================
# CLI Interface
# =============================================================================

def main():
    if not API_KEY:
        print("❌ Error: SILICONFLOW_API_KEY not set!")
        print("   Set with: export SILICONFLOW_API_KEY='your-key'")
        sys.exit(1)
    
    coder = Coder()
    
    print("🤖 Vyzorix Coder - Efficient AI Coding Assistant")
    print("="*50)
    print(f"Model: {coder.model}")
    print("="*50)
    print()
    print("Commands:")
    print("  review <files...>  - Review specific files")
    print("  ask <question>     - Ask a question")
    print("  fix <issue>        - Get fix suggestions")
    print("  refactor <task>    - Get refactoring help")
    print("  exit               - Quit")
    print()
    
    while True:
        try:
            user_input = input("> ").strip()
            
            if not user_input:
                continue
                
            if user_input.lower() in ['exit', 'quit', 'q']:
                print("Goodbye!")
                break
            
            # Parse command
            parts = user_input.split(maxsplit=1)
            cmd = parts[0].lower()
            args = parts[1] if len(parts) > 1 else ""
            
            if cmd == 'review':
                # Extract file paths
                files = args.split() if args else []
                if not files:
                    print("Usage: review <file1> <file2> ...")
                    continue
                print("\n🤖 Analyzing...")
                print(coder.review("Review these files and report issues", files))
                
            elif cmd == 'ask':
                if not args:
                    print("Usage: ask <question>")
                    continue
                print("\n🤖 Thinking...")
                print(coder.ask(args))
                
            elif cmd == 'fix':
                if not args:
                    print("Usage: fix <issue description>")
                    continue
                print("\n🤖 Analyzing...")
                print(coder.fix(args))
                
            elif cmd == 'refactor':
                if not args:
                    print("Usage: refactor <task>")
                    continue
                print("\n🤖 Analyzing...")
                print(coder.refactor(args))
                
            else:
                print(f"Unknown command: {cmd}")
                print("Commands: review, ask, fix, refactor, exit")
                
        except KeyboardInterrupt:
            print("\nGoodbye!")
            break


if __name__ == "__main__":
    main()
