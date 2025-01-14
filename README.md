# Interviewme🎯

---

## Basic Details




## Project Description

Developed an intelligent system that evaluates and scores resumes against job descriptions using advanced Natural Language Processing (NLP) techniques. The system employs Semantic Search to understand contextual relevance, Named Entity Recognition (NER) to extract key details like skills, experience, and qualifications, and Similarity Search to compare resume content with job requirements. The final score is calculated using a Weighted Average Model, ensuring key criteria such as skills, experience, and education are appropriately prioritized. Additionally, the model provides personalized feedback to candidates, highlighting strengths and areas for improvement. This project streamlines recruitment by automating resume screening and enhancing candidate-job alignment through data-driven insights.

---

## Technical Details

### Technologies/Components Used

**Languages:**
<div style="display: flex; align-items: center; gap: 20px; margin-bottom: 20px;">
    <img src="https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white" alt="Go" height="40"/>
    <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/javascript/javascript-original.svg" alt="JavaScript" height="40"/>
    <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/html5/html5-original.svg" alt="HTML5" height="40"/>
    <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/css3/css3-original.svg" alt="CSS3" height="40"/>
</div>


## **Frameworks/Libraries:**

<div style="display: flex; align-items: center; gap: 20px; margin-bottom: 20px;">
    <div style="text-align: center;">
        <img src="https://nextjs.org/static/favicon/favicon-32x32.png" alt="Next.js Logo" height="40"/>
        <p>Next.js</p>
    </div>
    <div style="text-align: center;">
        <img src="https://gofiber.io/assets/images/logo.svg" alt="Go Fiber" height="40"/>
        <p>Go Fiber</p>
    </div>
    <div style="text-align: center; font-size: 40px;">
        ☁️
        <p>Go Air</p>
    </div>
</div>



**Tools:**
<div style="display: flex; align-items: center; gap: 20px; margin-bottom: 20px;">
    <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/vscode/vscode-original.svg" alt="VSCode" height="40"/>
    <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/git/git-original.svg" alt="Git" height="40"/>
    <img src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png" alt="GitHub" height="40"/>
    <img src="https://www.gstatic.com/lamda/images/favicon_v1_150160cddff7f294ce30.svg" alt="Gemini" height="40"/>
    <img src="https://huggingface.co/front/assets/huggingface_logo-noborder.svg" alt="HuggingFace" height="40"/>
</div>

---

## Run

# 🚀 Project Setup Guide

A step-by-step guide to set up and run your project seamlessly.


## 📚 Prerequisites

Ensure the following tools are installed on your system:

- [Go (latest version)](https://go.dev/doc/install)  
- [Node.js (latest version)](https://nodejs.org/)  
- [npm (Node Package Manager)](https://www.npmjs.com/)  
- [Next.js](https://nextjs.org/docs/getting-started/installation)  



## 🛠️ Installation Steps

### 1. Clone the Repository

Clone the project to your local machine:

```bash
git [clone <your-repository-url>](https://github.com/nabeelnazeer/InterviewMe)
cd <your-repository>
```
## Install Go Dependencies
```bash
go mod tidy
```
## Install Node.js Dependencies
```bash
npm install
npm install next
```
## Environment Variables

Create a .env file in the backend directory and include the following:
```bash
HUGGING_FACE_API_KEY=your_hugging_face_api_key
GEMINI_API_KEY=your_gemini_api_key
PORT=8080


SESSIONS_DIR=processed_texts/sessions

```


Create a .env file in the frontend directory and include the following:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080

```

---

## Project Documentation

### Screenshots

![Home page](demo_folder/shot5.png)  


![score result](demo_folder/shot4.png)  

![Popup Interface](demo_folder/shot1.png)  


![Weird Expression](demo_folder/shot2.png)  


![Perfect Smile](demo_folder/shot3.png)  



## Team Contributions

- **Nabeel Nazeer**

---



# InterviewMe - Resume Scoring Platform


A modern web application that helps users score and evaluate resumes using various scoring techniques and algorithms.


### Frontend
- Next.js 13+ (React Framework)
- TypeScript
- Tailwind CSS
- shadcn/ui Components

### Backend
- Go 1.20+
- Fiber (Web Framework)
- GORM (ORM)


---

Made with ❤️
