@echo off
cd /d "%~dp0"
setlocal

:: Get current date and time in YYYY-MM-DD HH:MM:SS format
for /f "tokens=2 delims==" %%I in ('wmic os get localdatetime /value') do set datetime=%%I
set commit_message=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2% %datetime:~8,2%:%datetime:~10,2%:%datetime:~12,2%

:: Add all changes, including untracked files
git add -A

:: Commit and push
git commit -m "%commit_message%"
git push origin main

:: Optional: Pause to see output (remove if not needed)
echo.
echo Changes committed and pushed successfully!
pause
