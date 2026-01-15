### Telegram Uploader Workflow & Logic

#### Constraints and Limits
The system strictly adheres to Telegram's file size limitations to ensure accessibility for all users:
*   **Standard Limit:** 2 GB
*   **Premium Limit:** 4 GB
*   **System Policy:** To ensure all users can download the content easily, **the system enforces a hard limit of 2 GB per file**, regardless of the uploader's Premium status.

---

### 1. Preparation Phase
#### 1.1. Handling Non-Video Files (Context Preservation & Archiving)

To keep the content organized, non-video assets (PDFs, source code, archives) are separated from the video stream. The system replicates the original directory structure inside a staging folder to ensure users understand which module or section the files belong to.

Once organized, this folder is **compressed into ZIP archives**. If the total size of the assets exceeds the limit, the archive is automatically split into multiple parts to strictly respect the **2 GB limit per file**.

- **Original Path:** `introduction-1/course-setup/setup-instructions.pdf`
    
- **Staging Path:** `files/introduction-1/course-setup/setup-instructions.pdf`
    
- **Final Output:** `files.zip (or files.part1.zip, files.part2.zip if over 2 GB).`

#### 1.2. Video Splitting
Videos exceeding the **2 GB limit** are automatically split into multiple parts. The naming convention appends the part number to the original filename:

*   **Original:** `golang-routines.mp4`
*   **Split:** `golang-routines-part-1.mp4`, `golang-routines-part-2.mp4`, etc.

#### 1.3 Create an `index.md`

This content is sent as the **final message** in the channel and is automatically **pinned**.

**Index Example:**
```markdown 
⚠️ **Attention** ⚠️
 
Click here to view the Menu. 
Use the hashtags to jump directly to specific videos.

📂 **Files**
#Doc001

🔹 **01 - Introduction**
#F001 #F002 #F003 #F004

🔹 **02 - PyTorch Fundamentals**
#F005 #F006 #F007 #F008 #F009 #F010 #F011 #F012 

🔹 **03 - PyTorch Workflow**
#F037 #F038 #F039 #F040 #F041 #F042 #F043
```

#### 1.4. Metadata & Captions
Every uploaded video includes a structured caption to facilitate search and navigation. The format includes:
1.  **Navigation Hashtag:** A unique ID (e.g., `#D033`).
2.  **Video Title:** Preceded by its sequence number.
3.  **Module Name:** The parent category of the video.

**Caption Template:**
```markdown
#D033 8 - Creating the first commit
Module 03 - Code versioning and hosting (Git and Github)
```

#### 1.5. Upload Queue
The system generates an upload queue based on the file system hierarchy. It processes folders and files in **alphanumeric ascending order**.

**Recommended Directory Structure:**
For the best results, ensure your source folders are numbered:
```text
├── 1. Introduction
│   ├── 1. Welcome and Support
│   │   ├── 1. first-class.mp4
│   │   └── description.html
│   └── 2. The initial setup
│       └── 1. how-to-setup.mp4
```

---

### 2. Execution Phase

#### 2.1. Upload Sequence
To ensure the user experience is logical, the upload order is strictly defined:
1.  **Assets:** Upload all non-video files (replicated structure).
2.  **Videos:** Upload all video files (in queue order).
3.  **Index:** Generate and send the navigation menu.

#### 2.2. Finalize
- Send the `index.md` content (which was created in step 1.3) as the final message.    
- Pin the message.