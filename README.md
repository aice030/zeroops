# ZeroOps 

## Prerequisites

- Python 3.11 or higher
- pip package manager

## Installation

### macOS Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/zeroops.git
   cd zeroops
   ```

2. **Create and activate virtual environment**
   ```bash
   # Create virtual environment
   python3 -m venv venv
   
   # Activate virtual environment
   source venv/bin/activate
   ```

3. **Install dependencies**
   ```bash
   pip install -r requirements.txt
   ```

4. **Verify installation**
   ```bash
   python -c "import agents; print('ZeroOps installed successfully!')"
   ```

### Windows Installation

1. **Clone the repository**
   ```cmd
   git clone https://github.com/yourusername/zeroops.git
   cd zeroops
   ```

2. **Create and activate virtual environment**
   ```cmd
   # Create virtual environment
   python -m venv venv
   
   # Activate virtual environment (Command Prompt)
   venv\Scripts\activate.bat
   
   # OR activate virtual environment (PowerShell)
   venv\Scripts\Activate.ps1
   ```

3. **Install dependencies**
   ```cmd
   pip install -r requirements.txt
   ```

4. **Verify installation**
   ```cmd
   python -c "import agents; print('ZeroOps installed successfully!')"
   ```

## Dependencies

The main dependencies include:
- `openai>=1.99.0` - OpenAI API client
- `transformers>=4.55.0` - Hugging Face transformers
- `numpy>=2.3.0` - Numerical computing
- `requests>=2.31.0` - HTTP library
- `rich>=14.1.0` - Rich terminal formatting