'use client';
import { useState, useRef, useEffect } from 'react';
import { FiUpload, FiTrash2 } from 'react-icons/fi';
import { FaCode, FaDatabase, FaCloud, FaMobile, FaDesktop, FaRobot, FaChartLine, FaCuttlefish } from 'react-icons/fa';
import { useRouter } from 'next/navigation';

const predefinedJobs = [
  {
    id: 'frontend',
    title: 'UI/UX designer',
    icon: FaCode,
    color: '#10B981',
    description: `Job Title: UX/UI Designer

Role Summary: We are in search of a UX/UI Designer who is passionate about improving user experience by creating intuitive, user-friendly design solutions. This position is suitable for those who are at the early stage of their career and have a deep interest in interactive design.

Responsibilities:
- Design and implement user interfaces for different digital platforms.
- Collaborate with the product and engineering team to define and implement innovative solutions for the product direction, visuals, and experience.
- Develop wireframes, user flows, and prototypes to effectively communicate interaction and design ideas.
- Conduct user research and evaluate user feedback to optimize the design.
- Establish and promote design guidelines, best practices, and standards.

Requirements:
- Degree in Design, Computer Science or a related field.
- 0-3 years of experience in UX/UI design.
- Proficiency in graphic design software including Adobe Photoshop, Adobe Illustrator, and other visual design tools.
- Familiarity with HTML, CSS, and JavaScript for rapid prototyping.
- Strong visual design skills with a good understanding of user-system interaction.
- Ability to solve problems creatively and effectively.
- Excellent verbal and written communication skills.
- Up-to-date with the latest UI trends, techniques, and technologies.`
  },
  {
    id: 'backend',
    title: 'Backend Developer',
    icon: FaDatabase,
    color: '#6366F1',
    description: `# **Job Title:** Backend Developer  

**Location:** [Remote/On-Site/Hybrid]  
**Company:** [Company Name]  
**Employment Type:** [Full-time/Part-time/Contract]  
**Experience Level:** [Junior/Mid-Level/Senior]  

---

## **About Us:**  
[Company Name] is a *[brief company description, e.g., leading tech company specializing in SaaS solutions for businesses worldwide]*. We are passionate about building scalable, high-performance applications that deliver exceptional user experiences.  

---

## **Role Overview:**  
We are seeking a skilled **Backend Developer** to join our dynamic development team. In this role, you will be responsible for designing, implementing, and maintaining robust backend systems that power our web and mobile applications.  

---

## **Key Responsibilities:**  
- Design, develop, and maintain server-side logic, ensuring high performance and responsiveness to API requests.  
- Develop and maintain **RESTful** and **GraphQL APIs**.  
- Optimize backend processes for scalability, reliability, and efficiency.  
- Collaborate with front-end developers, product managers, and other team members to deliver seamless integrations.  
- Manage databases, including schema design, query optimization, and data integrity.  
- Write clean, well-documented, and reusable code following industry best practices.  
- Implement security and data protection protocols.  
- Perform debugging, troubleshooting, and root cause analysis of production issues.  
- Participate in code reviews and knowledge-sharing sessions.  

---

## **Required Skills and Qualifications:**  
- Proficiency in backend programming languages such as **Python (Django/Flask)**, **Node.js (Express.js)**, **Java (Spring Boot)**, or **GoLang**.  
- Experience with database technologies such as **MySQL**, **PostgreSQL**, **MongoDB**, or **Redis**.  
- Familiarity with cloud platforms such as **AWS**, **Azure**, or **Google Cloud Platform (GCP)**.  
- Strong understanding of **RESTful APIs** and **microservices architecture**.  
- Proficiency in version control tools like **Git**.  
- Knowledge of **CI/CD pipelines** and deployment automation.  
- Solid understanding of security best practices and authentication mechanisms (e.g., **OAuth**, **JWT**).  
- Strong problem-solving and analytical skills.  
- Good communication and teamwork skills.  

---

## **Nice-to-Have:**  
- Experience with containerization tools like **Docker** and orchestration tools like **Kubernetes**.  
- Familiarity with message brokers like **RabbitMQ** or **Kafka**.  
- Knowledge of **GraphQL APIs**.  
- Prior experience with performance testing and monitoring tools.  

---

## **Benefits:**  
- Competitive salary and performance-based incentives.  
- Flexible working hours and remote work options.  
- Health insurance and wellness programs.  
- Opportunities for professional growth and development.  
- Collaborative and inclusive work culture.  

---

## **How to Apply:**  
Submit your **resume** and a **cover letter** explaining why you're the perfect fit for this role to **[email@example.com]** or apply directly on our website: **[company website/careers page]**.  

We look forward to hearing from you! 🚀  
`
  },
  {
    id: 'cloud',
    title: 'Cloud Engineer',
    icon: FaCloud,
    color: '#EC4899',
    description: `Looking for a Cloud Engineer with AWS/Azure expertise, Infrastructure as Code experience, 
    and strong DevOps practices. Knowledge of containerization and orchestration required.`
  },
  {
    id: 'mobile',
    title: 'Mobile Developer',
    icon: FaMobile,
    color: '#F59E0B',
    description: `Mobile Developer position requiring React Native/Flutter experience, 
    knowledge of mobile UI/UX principles, and app deployment processes.`
  },
  {
    id: 'fullstack',
    title: 'Full Stack',
    icon: FaDesktop,
    color: '#8B5CF6',
    description: `Full Stack Developer role requiring expertise in both frontend and backend technologies, 
    database management, and modern web development practices.`
  },
  {
    id: 'ai',
    title: 'AI Engineer',
    icon: FaRobot,
    color: '#EF4444',
    description: `AI Engineer position focusing on machine learning model development, 
    deep learning frameworks, and ML ops. Experience with PyTorch/TensorFlow required.`
  },
  {
    id: 'data',
    title: 'Data Scientist',
    icon: FaChartLine,
    color: '#14B8A6',
    description: `Data Scientist role requiring expertise in statistical analysis, 
    machine learning, and data visualization. Python and SQL proficiency needed.`
  },
  {
    id: 'custom',
    title: 'Custom Job Description',
    icon: FaCuttlefish,
    color: '#D946EF',
    description: `lajavathiye`
  }
];

export default function CVScoring() {
  const fileInputRef = useRef(null);
  const [uploadedCV, setUploadedCV] = useState(null);
  const [jobDescription, setJobDescription] = useState('');
  const [modelScores, setModelScores] = useState({});
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState(null);
  const [processingModels, setProcessingModels] = useState({});
  const [selectedJob, setSelectedJob] = useState(null);
  const [isProcessingJob, setIsProcessingJob] = useState(false);
  const [jobAnalysis, setJobAnalysis] = useState(null);
  const [isPreprocessing, setIsPreprocessing] = useState(false);
  const [preprocessedData, setPreprocessedData] = useState(null);
  const [scoringResults, setScoringResults] = useState(null);
  const [isScoring, setIsScoring] = useState(false);
  const [isClearing, setIsClearing] = useState(false);
  const router = useRouter();

  const aiModels = [
    { id: 'gpt4', name: 'GPT-4 Analysis', color: '#10B981' },
    { id: 'bert', name: 'BERT Matcher', color: '#6366F1' },
    { id: 'llama', name: 'LLaMA Evaluator', color: '#EC4899' },
    { id: 'custom', name: 'Custom AI Model', color: '#F59E0B' }
  ];

  const handleFileUpload = async (event) => {
    const file = event.target.files[0];
    if (file) {
      setIsUploading(true);
      setError(null);
      setPreprocessedData(null);

      try {
        // First, upload the file
        const uploadFormData = new FormData();
        uploadFormData.append('file', file);
        
        const uploadResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/upload`, {
          method: 'POST',
          body: uploadFormData,
        });

        if (!uploadResponse.ok) {
          throw new Error('Failed to upload file');
        }

        const uploadResult = await uploadResponse.json();

        // Then, preprocess the uploaded file
        const preprocessFormData = new FormData();
        preprocessFormData.append('resume', file);
        
        const preprocessResponse = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/preprocess`, {
          method: 'POST',
          body: preprocessFormData,
        });

        if (!preprocessResponse.ok) {
          throw new Error('Failed to process CV');
        }

        const result = await preprocessResponse.json();
        // Add filename to preprocessed data
        result.filename = uploadResult.filename;
        
        setUploadedCV(file);
        setPreprocessedData(result);

      } catch (err) {
        setError(err.message || 'Failed to process CV. Please try again.');
        console.error('Processing error:', err);
      } finally {
        setIsUploading(false);
      }
    }
  };

  const handleDelete = async () => {
    try {
      // Delete the file from server
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/delete?filename=${uploadedCV.name}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Delete failed');
      }

      // Only clear states after successful deletion
      setUploadedCV(null);
      setModelScores({});
      setError(null);
      setPreprocessedData(null);
      setProcessingModels({});
      
      // Reset the file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }

    } catch (err) {
      setError('Failed to delete CV. Please try again.');
      console.error('Delete error:', err);
    }
  };

  const processWithModel = async (modelId) => {
    setProcessingModels(prev => ({ ...prev, [modelId]: true }));
    try {
      // Simulate API call to specific model endpoint
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/analyze/${modelId}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cv: uploadedCV?.name, jobDescription }),
      });
      
      if (!response.ok) throw new Error(`${modelId} analysis failed`);
      
      const result = await response.json();
      setModelScores(prev => ({ ...prev, [modelId]: result }));
    } catch (err) {
      setError(`Failed to process with ${modelId}. Please try again.`);
    } finally {
      setProcessingModels(prev => ({ ...prev, [modelId]: false }));
    }
  };

  const handleJobSelect = async (job) => {
    setSelectedJob(job.id);
    setJobDescription(job.description);
    setJobAnalysis(null);
    
    setIsProcessingJob(true);
    try {
      const timestamp = Math.floor(Date.now() / 1000);
      const jobId = `job_${timestamp}`; // Create clean job ID

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/preprocess-job`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          description: job.description,
          job_id: jobId
        }),
      });

      if (!response.ok) throw new Error('Failed to process job description');
      
      const result = await response.json();
      
      // Store the clean filename
      const jobData = {
        ...result,
        id: jobId,
        filename: `job_${timestamp}.json` // Store the correct filename format
      };
      
      setJobAnalysis(jobData);
      localStorage.setItem('jobFilename', jobData.filename); // Store for later use

    } catch (err) {
      setError('Failed to process job description: ' + err.message);
      console.error('Job processing error:', err);
    } finally {
      setIsProcessingJob(false);
    }
  };

  const handleScore = async () => {
    if (!uploadedCV || !jobAnalysis) {
      setError('Please upload a CV and select a job description first');
      return;
    }

    setIsScoring(true);
    setError(null);

    try {
      console.log('Starting CV scoring process...');
      
      // Clean job ID by removing any 'job_' prefix
      const cleanJobId = jobAnalysis.id.replace(/^job_+/, '');
      
      const requestData = {
        resume_id: preprocessedData.id,
        job_id: cleanJobId
      };

      console.log('Sending scoring request with:', requestData);

      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/score-resume`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData),
      });

      if (!response.ok) {
        throw new Error(`Score API failed: ${response.status} ${response.statusText}`);
      }

      const results = await response.json();
      console.log('Received scoring results:', results);
      
      // Store the correct IDs format
      const dataToStore = {
        scores: results,
        resumeData: preprocessedData,
        jobData: jobAnalysis,
        resumeId: preprocessedData.id, // Store the correct format
        jobId: cleanJobId,
        resumeFilename: preprocessedData.filename,
        jobFilename: jobAnalysis.filename
      };

      localStorage.setItem('scoringResults', JSON.stringify(dataToStore));
      console.log('Stored data with correct IDs:', dataToStore);

      await router.push('/analysis');

    } catch (err) {
      console.error('Scoring error:', err);
      setError(`Failed to score CV: ${err.message}`);
      setIsScoring(false);
    }
  };

  const handleClear = async () => {
    setIsClearing(true);
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/clear`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error('Failed to clear files');
      }

      // Reset all states
      setUploadedCV(null);
      setJobDescription('');
      setSelectedJob(null);
      setJobAnalysis(null);
      setPreprocessedData(null);
      setScoringResults(null);
      setError(null);
      
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }

    } catch (err) {
      setError('Failed to clear files: ' + err.message);
    } finally {
      setIsClearing(false);
    }
  };

  const RequirementsList = ({ title, items, color }) => (
    items && items.length > 0 && (
      <div className="mb-3">
        <h4 className="text-sm font-semibold mb-1" style={{ color }}>
          {title}
        </h4>
        <ul className="list-disc list-inside text-sm text-gray-300">
          {items.map((item, idx) => (
            <li key={idx}>{item}</li>
          ))}
        </ul>
      </div>
    )
  );

  const ExperienceSection = ({ experience, color }) => (
    experience && (
      <div className="mb-3">
        <h4 className="text-sm font-semibold mb-1" style={{ color }}>
          Experience Requirements
        </h4>
        <div className="text-sm text-gray-300">
          <p>Level: {experience.level}</p>
          <p>Minimum Years: {experience.min_years}</p>
          {experience.areas && experience.areas.length > 0 && (
            <div>
              <p>Areas:</p>
              <ul className="list-disc list-inside pl-4">
                {experience.areas.map((area, idx) => (
                  <li key={idx}>{area}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    )
  );

  const EducationSection = ({ education, color }) => (
    education && (
      <div className="mb-3">
        <h4 className="text-sm font-semibold mb-1" style={{ color }}>
          Education Requirements
        </h4>
        <div className="text-sm text-gray-300">
          <p>Degree: {education.degree}</p>
          {education.fields && (
            <div>
              <p>Fields:</p>
              <ul className="list-disc list-inside pl-4">
                {education.fields.map((field, idx) => (
                  <li key={idx}>{field}</li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    )
  );

  const ScoringResults = ({ results }) => {
    if (!results) return null;

    // Separate important scores from other scores
    const importantScores = [
      { label: 'Overall Score', value: results.overall_score, color: '#10B981', highlight: true },
      { label: 'Technical Skills', value: results.detailed_scores.technical_skills, color: '#6366F1', highlight: true },
      { label: 'Skills Match', value: results.skills_match, color: '#EC4899', highlight: true },
    ];

    const otherScores = [
      { label: 'Experience', value: results.experience_match, color: '#F59E0B' },
      { label: 'Education', value: results.education_match, color: '#8B5CF6' },
      ...Object.entries(results.detailed_scores || {})
        .filter(([key]) => key !== 'technical_skills')
        .map(([key, value]) => ({
          label: key.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' '),
          value: value,
          color: '#8B5CF6'
        }))
    ];

    return (
      <div className="bg-gray-800 rounded-xl p-6 mt-6">
        <h3 className="text-xl font-bold text-white mb-6">Analysis Results</h3>
        
        {/* Important Scores with Enhanced Styling */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          {importantScores.map((score, index) => (
            <div 
              key={index}
              className="bg-gray-700 rounded-lg p-6 text-center transform transition-all duration-500 hover:scale-105 relative overflow-hidden"
              style={{ 
                borderLeft: `4px solid ${score.color}`,
                animation: `fadeIn 0.5s ease-out ${index * 0.1}s forwards`,
                opacity: 0
              }}
            >
              {/* Add highlight effect */}
              <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/5 to-transparent"/>
              <div className="relative z-10">
                <div className="text-3xl font-bold text-white mb-2">
                  {score.value}%
                </div>
                <div className="text-sm text-gray-300 font-semibold">{score.label}</div>
              </div>
            </div>
          ))}
        </div>

        {/* Other Scores with Regular Styling */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {otherScores.map((score, index) => (
            <div 
              key={index}
              className="bg-gray-700/50 rounded-lg p-3 text-center"
              style={{ 
                borderTop: `4px solid ${score.color}`,
                animation: `fadeIn 0.5s ease-out ${(index + importantScores.length) * 0.1}s forwards`,
                opacity: 0
              }}
            >
              <div className="text-xl font-bold text-gray-300">
                {score.value}%
              </div>
              <div className="text-xs text-gray-400">{score.label}</div>
            </div>
          ))}
        </div>

        {/* Feedback Section - unchanged */}
        {results.feedback && results.feedback.length > 0 && (
          <div className="mt-6" style={{ animation: 'fadeIn 0.5s ease-out 0.8s forwards', opacity: 0 }}>
            <h4 className="text-lg font-semibold text-white mb-3">Improvement Areas</h4>
            <div className="space-y-2">
              {results.feedback.map((item, index) => (
                <div 
                  key={index}
                  className="bg-gray-700 p-3 rounded-lg text-gray-300 text-sm transform transition-all duration-500 hover:scale-[1.02]"
                  style={{ 
                    borderLeft: '4px solid #F59E0B',
                    animation: `slideIn 0.5s ease-out ${index * 0.1 + 0.9}s forwards`,
                    opacity: 0,
                    transform: 'translateX(-20px)'
                  }}
                >
                  {item}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  };

  // Add this style block at the end of your component before the final export
  const styles = `
    @keyframes fadeIn {
      from { opacity: 0; }
      to { opacity: 1; }
    }
    @keyframes slideIn {
      from { 
        opacity: 0;
        transform: translateX(-20px);
      }
      to { 
        opacity: 1;
        transform: translateX(0);
      }
    }
  `;

  // Add this right before the final return statement
  useEffect(() => {
    // Inject the styles
    const styleSheet = document.createElement("style");
    styleSheet.innerText = styles;
    document.head.appendChild(styleSheet);
    return () => {
      document.head.removeChild(styleSheet);
    };
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      // Resume preprocessing
      const resumeResponse = await axios.post(`${process.env.NEXT_PUBLIC_API_URL}/preprocess`, formData);
      localStorage.setItem('session_id_resume', resumeResponse.data.session_id);
      
      // Job preprocessing
      const jobResponse = await axios.post(`${process.env.NEXT_PUBLIC_API_URL}/preprocess-job`, {
        description: jobDescription
      });
      localStorage.setItem('session_id_job', jobResponse.data.session_id_job);
  
      // Store the full responses
      const scoringResults = {
        scores: resumeResponse.data.scores,
        resumeData: resumeResponse.data,
        jobData: jobResponse.data,
        resumeId: resumeResponse.data.id,
        jobId: jobResponse.data.id,
        resumeResponse,
        jobResponse
      };
      
      localStorage.setItem('scoringResults', JSON.stringify(scoringResults));
      
      // Log stored session IDs
      console.log('Stored session IDs:', {
        resume: resumeResponse.data.session_id,
        job: jobResponse.data.session_id_job
      });
  
      router.push('/analysis');
    } catch (error) {
      console.error('Error during preprocessing:', error);
      setError('Failed to process files');
    } finally {
      setLoading(false);
    }
  };

  const handleJobDescriptionSubmit = async () => {
    setIsLoading(true);
    try {
      const response = await axios.post(
        `${process.env.NEXT_PUBLIC_API_URL}/preprocess-job`,
        { description: jobDescription }
      );
  
      // Store the job filename for later use
      if (response.data.filename) {
        localStorage.setItem('jobFilename', response.data.filename);
      }
  
      // Store full response for scoring
      localStorage.setItem('jobResponse', JSON.stringify(response.data));
  
      // Rest of your existing code...
    } catch (error) {
      // ...existing error handling...
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 p-8" suppressHydrationWarning>
      <div className="max-w-7xl mx-auto space-y-8">
        <h1 className="text-4xl font-bold text-white">
          Resume Evaluator
          <span className="text-green-500">.</span>
        </h1>
        
        {/* Job Selection Section */}
        <div className="mb-8">
          <h2 className="text-xl text-white mb-4">Select Job Position</h2>
          <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-4">
            {predefinedJobs.map((job) => (
              <button
                key={job.id}
                onClick={() => handleJobSelect(job)}
                className={`p-4 rounded-lg border-2 transition-all hover:scale-105 flex flex-col items-center ${
                  selectedJob === job.id ? 'border-opacity-100' : 'border-opacity-50'
                }`}
                style={{ borderColor: job.color, backgroundColor: selectedJob === job.id ? `${job.color}20` : 'transparent' }}
              >
                <job.icon className="w-8 h-8 mb-2" style={{ color: job.color }} />
                <span className="text-white text-sm text-center">{job.title}</span>
              </button>
            ))}
          </div>
        </div>

        {/* Upload and Job Description Section */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          <div className="bg-gray-800 p-6 rounded-lg border-2 border-opacity-50" style={{ borderColor: '#10B981' }}>
            <h2 className="text-xl text-white mb-4">Upload CV</h2>
            <div className="flex items-center justify-center w-full">
              <label className={`flex flex-col items-center justify-center w-full h-48 border-2 border-gray-600 border-dashed rounded-lg cursor-pointer hover:bg-gray-700 ${isUploading || isPreprocessing ? 'opacity-50' : ''}`}>
                <div className="flex flex-col items-center justify-center pt-5 pb-6">
                  <FiUpload className="w-10 h-10 text-gray-400 mb-3" />
                  <p className="text-sm text-gray-400">
                    {isUploading ? "Uploading..." : 
                     isPreprocessing ? "Extracting information..." :
                     uploadedCV ? uploadedCV.name : 
                     "Click to upload CV"}
                  </p>
                </div>
                <input 
                  ref={fileInputRef}
                  type="file" 
                  className="hidden" 
                  onChange={handleFileUpload} 
                  accept=".pdf,.doc,.docx"
                  disabled={isUploading || isPreprocessing} 
                />
              </label>
            </div>
            {preprocessedData && (
              <div className="mt-4 bg-gray-800 rounded-lg p-4">
                <h3 className="font-semibold mb-2">Extracted Information:</h3>
                <div className="bg-gray-700 p-3 rounded text-sm space-y-2">
                  {/* Personal Information */}
                  <div className="space-y-2 pb-2 border-b border-gray-600">
                    <div>
                      <span className="text-gray-400">Name:</span>{' '}
                      {(preprocessedData.result?.entities?.name || preprocessedData?.entities?.name) ?? 'Not found'}
                    </div>
                    <div>
                      <span className="text-gray-400">Email:</span>{' '}
                      {(preprocessedData.result?.entities?.email?.[0] || preprocessedData?.entities?.email?.[0]) ?? 'Not found'}
                    </div>
                    <div>
                      <span className="text-gray-400">Phone:</span>{' '}
                      {(preprocessedData.result?.entities?.phone || preprocessedData?.entities?.phone) ?? 'Not found'}
                    </div>
                  </div>

                  {/* Technical Skills */}
                  {(preprocessedData.result?.technical_skills || preprocessedData?.technical_skills)?.length > 0 && (
                    <div className="pt-2">
                      <span className="text-gray-400 block mb-2">Technical Skills:</span>
                      <div className="flex flex-wrap gap-2">
                        {(preprocessedData.result?.technical_skills || preprocessedData?.technical_skills).map((skill, index) => (
                          <span 
                            key={index} 
                            className="px-2 py-1 bg-gray-600 rounded text-xs 
                                     border border-gray-500 text-gray-200"
                          >
                            {skill}
                          </span>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Education */}
                  {(preprocessedData.result?.education || preprocessedData?.education)?.length > 0 && (
                    <div className="pt-2 border-t border-gray-600">
                      <span className="text-gray-400 block mb-2">Education:</span>
                      <div className="space-y-2">
                        {(preprocessedData.result?.education || preprocessedData?.education).map((edu, index) => (
                          <div 
                            key={index} 
                            className="text-sm bg-gray-600/50 p-2 rounded border border-gray-600"
                          >
                            {edu}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Experience */}
                  {preprocessedData.result?.experience && preprocessedData.result.experience.length > 0 && (
                    <div className="pt-2 border-t border-gray-600">
                      <span className="text-gray-400 block mb-2">Experience:</span>
                      <div className="space-y-3">
                        {preprocessedData.result.experience.map((exp, index) => (
                          <div 
                            key={index} 
                            className="bg-gray-600/50 p-2 rounded border border-gray-600"
                          >
                            <div className="font-medium text-white">{exp.title}</div>
                            <div className="text-sm text-gray-300">{exp.company}</div>
                            <div className="text-xs text-gray-400">{exp.duration}</div>
                            {exp.description && (
                              <div className="text-sm text-gray-300 mt-1">{exp.description}</div>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}
            {error && (
              <p className="mt-2 text-red-400 text-sm">{error}</p>
            )}
            {uploadedCV && (
              <button
                onClick={handleDelete}
                className="mt-4 flex items-center text-red-400 hover:text-red-300"
              >
                <FiTrash2 className="mr-2" /> Remove CV
              </button>
            )}
          </div>

          <div className="bg-gray-800 p-6 rounded-lg border-2 border-opacity-50" style={{ borderColor: '#6366F1' }}>
            <h2 className="text-xl text-white mb-4">Job Description</h2>
            <textarea
              className="w-full h-48 bg-gray-700 text-white rounded-lg p-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Select a job position or enter custom job description..."
              value={jobDescription}
              onChange={(e) => setJobDescription(e.target.value)}
              disabled={isProcessingJob}
            />
            {isProcessingJob && (
              <div className="mt-2 text-blue-400">
                <span className="animate-pulse">Extracting requirements...</span>
              </div>
            )}
            {jobAnalysis?.requirements && (
              <div className="mt-4 text-white">
                <h3 className="font-semibold mb-3">Job Requirements Analysis:</h3>
                <div className="bg-gray-700 p-4 rounded space-y-4">
                  {jobAnalysis.requirements.skills?.length > 0 && (
                    <RequirementsList 
                      title="Required Skills" 
                      items={jobAnalysis.requirements.skills} 
                      color="#10B981"
                    />
                  )}
                  {jobAnalysis.requirements.experience && (
                    <ExperienceSection 
                      experience={jobAnalysis.requirements.experience} 
                      color="#6366F1"
                    />
                  )}
                  {jobAnalysis.requirements.education && (
                    <EducationSection 
                      education={jobAnalysis.requirements.education} 
                      color="#EC4899"
                    />
                  )}
                  {jobAnalysis.requirements.responsibilities?.length > 0 && (
                    <RequirementsList 
                      title="Key Responsibilities" 
                      items={jobAnalysis.requirements.responsibilities} 
                      color="#F59E0B"
                    />
                  )}
                </div>
              </div>
            )}
          </div>
        </div>

        {error && (
          <div className="mt-4 p-4 bg-red-900/50 border border-red-500 rounded-lg text-red-200">
            {error}
          </div>
        )}

        {/* Add Score and Clear buttons */}
        <div className="flex gap-4 mt-8">
          <button
            onClick={handleScore}
            disabled={!uploadedCV || !jobAnalysis || isScoring}
            className={`px-6 py-2 rounded-lg text-white font-medium transition-all duration-300 ${
              !uploadedCV || !jobAnalysis || isScoring
                ? 'bg-gray-600 cursor-not-allowed'
                : 'bg-green-600 hover:bg-green-700 hover:scale-105'
            }`}
          >
            {isScoring ? 'Analyzing...' : 'Score CV'}
          </button>

          <button
            onClick={handleClear}
            disabled={isClearing}
            className="px-6 py-2 rounded-lg text-white font-medium bg-red-600 hover:bg-red-700"
          >
            {isClearing ? 'Clearing...' : 'Clear All Files'}
          </button>
        </div>

        {/* Add Scoring Results component */}
        <ScoringResults results={scoringResults} />
      </div>
    </div>
  );
}
