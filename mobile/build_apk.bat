@echo off
echo Building ProjectX DAO Mobile APK...
echo.

echo Step 1: Cleaning previous builds...
flutter clean

echo Step 2: Getting dependencies...
flutter pub get

echo Step 3: Running code generation...
flutter packages pub run build_runner build --delete-conflicting-outputs

echo Step 4: Building APK (Debug)...
flutter build apk --debug

echo Step 5: Building APK (Release)...
flutter build apk --release

echo.
echo Build completed!
echo Debug APK: build\app\outputs\flutter-apk\app-debug.apk
echo Release APK: build\app\outputs\flutter-apk\app-release.apk
echo.
pause