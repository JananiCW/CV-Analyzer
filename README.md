<h1 align = "center"> 🎯 AI CV Analyzer </h1>

An AI-powered web application that analyzes CVs (PDF format) and provides professional feedback including candidate summary, key skills, strengths, and improvement suggestions.

The system extracts text from uploaded CVs, sends it to Google Gemini AI for analysis, and stores the results in a MongoDB database for later viewing.

---

## 🚀 Features

* 📄 Upload CV in **PDF format**
* 🤖 AI-powered CV analysis using **Google Gemini**
* 📊 Generates:

  * Candidate summary
  * Key skills detected
  * Experience level
  * Suitable job roles
  * Strengths
  * Areas for improvement
  * Overall score
* 🗂 Stores previous analyses in **MongoDB**
* 👁 View past CV analysis results
* 🗑 Delete previous analyses
* 🎨 Clean and modern UI

---

## 🏗 Tech Stack

### Frontend

* React
* CSS

### Backend

* Go (Golang)

### Database

* MongoDB

### AI Service

* Google Gemini API

### Other Libraries

* PDF text extraction library
* dotenv for environment variables

---

## ⚙️ System Architecture

User uploads CV → React frontend → Go backend → PDF text extraction → Gemini AI analysis → MongoDB storage → Results displayed in UI

##Screenshots

<img width="1116" height="868" alt="image" src="https://github.com/user-attachments/assets/e047d9a5-2f61-4846-a6a6-a05d6a6d203f" />
<br/>
<img width="1441" height="872" alt="image" src="https://github.com/user-attachments/assets/df6cd06c-5ac2-4fb5-886f-fe7f7b5c901c" />
<br/>
<img width="1242" height="862" alt="image" src="https://github.com/user-attachments/assets/3e75ae2e-1e32-4ec9-ac54-a56475395244" />

