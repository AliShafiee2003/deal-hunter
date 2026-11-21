import os
import subprocess
from pathlib import Path


def extract_audio(
    input_file: str | Path,
    output_format: str = "mp3",  # 'mp3' or 'm4a'
    bitrate: str = "64k",
) -> Path:
    """
    Extracts audio from video with minimal size and high voice quality.

    - If output_format == 'm4a': Direct stream copy (-c:a copy), 100% original quality, zero re-encoding.
    - If output_format == 'mp3': Re-encoded with libmp3lame, 64kbps mono (ideal for speech/lectures).
    """
    input_path = Path(input_file).resolve()
    if not input_path.exists():
        raise FileNotFoundError(f"فایل ورودی یافت نشد: {input_path}")

    out_ext = f".{output_format.lower().lstrip('.')}"
    output_path = input_path.with_suffix(out_ext)

    print(f"در حال پردازش: {input_path.name}")

    if output_format.lower() in ["m4a", "aac"]:
        # کپی مستقیم استریم بدون افت کیفیت و با حداکثر سرعت
        cmd = [
            "ffmpeg",
            "-y",
            "-i",
            str(input_path),
            "-vn",
            "-c:a",
            "copy",
            str(output_path),
        ]
    else:
        # تبدیل به MP3 با فشرده‌سازی استاندارد و شفافیت عالی برای صوت و سخنرانی
        cmd = [
            "ffmpeg",
            "-y",
            "-i",
            str(input_path),
            "-vn",
            "-c:a",
            "libmp3lame",
            "-b:a",
            bitrate,
            "-ac",
            "1",
            str(output_path),
        ]

    result = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8")
    if result.returncode != 0:
        raise RuntimeError(f"خطا در اجرای ffmpeg:\n{result.stderr}")

    size_mb = output_path.stat().st_size / (1024 * 1024)
    print(f"انجام شد: {output_path.name} | حجم: {size_mb:.2f} MB\n")
    return output_path


def main():
    downloads_dir = Path(r"C:\Users\alish\Downloads")
    files = [
        "کلاس مجازي رايانش ابري [4051-1920076-01]_0.mp4",
        "کلاس مجازي رايانش ابري [4051-1920076-01]_1.mp4",
    ]

    for fname in files:
        video_path = downloads_dir / fname
        if video_path.exists():
            # استخراج نسخه MP3 (سازگاری عمومی بالا در همه پلیرها و موبایل)
            extract_audio(video_path, output_format="mp3", bitrate="64k")
            # استخراج نسخه M4A (کیفیت 100٪ بدون فشرده‌سازی مجدد و حجم اندک)
            extract_audio(video_path, output_format="m4a")
        else:
            print(f"فایل یافت نشد: {video_path}")


if __name__ == "__main__":
    main()
