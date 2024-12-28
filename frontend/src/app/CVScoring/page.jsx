'use client';
import { useState } from 'react';
import { FiUpload, FiTrash2, FiCpu } from 'react-icons/fi';
import { FaCode, FaDatabase, FaCloud, FaMobile, FaDesktop, FaRobot, FaChartLine, FaShieldAlt, FaCuttlefish } from 'react-icons/fa';

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
    description: `Seeking a Backend Developer with strong Python/Node.js skills and experience with REST APIs, 
    database design, and server architecture. Knowledge of microservices and cloud platforms preferred.`
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
        
        const uploadResponse = await fetch('http://localhost:8080/api/upload', {
          method: 'POST',
          body: uploadFormData,
        });

        if (!uploadResponse.ok) {
          throw new Error('Failed to upload file');
        }

        // Then, preprocess the uploaded file
        const preprocessFormData = new FormData();
        preprocessFormData.append('resume', file);
        
        const preprocessResponse = await fetch('http://localhost:8080/api/preprocess', {
          method: 'POST',
          body: preprocessFormData,
        });

        if (!preprocessResponse.ok) {
          throw new Error('Failed to process CV');
        }

        const result = await preprocessResponse.json();
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
      const response = await fetch(`http://localhost:8080/api/delete?filename=${uploadedCV.name}`, {
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

    } catch (err) {
      setError('Failed to delete CV. Please try again.');
      console.error('Delete error:', err);
    }
  };

  const processWithModel = async (modelId) => {
    setProcessingModels(prev => ({ ...prev, [modelId]: true }));
    try {
      // Simulate API call to specific model endpoint
      const response = await fetch(`http://localhost:8080/api/analyze/${modelId}`, {
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
      const response = await fetch('http://localhost:8080/api/preprocess-job', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ description: job.description }),
      });

      if (!response.ok) throw new Error('Failed to process job description');
      
      const result = await response.json();
      setJobAnalysis(result);
    } catch (err) {
      setError('Failed to process job description');
      console.error(err);
    } finally {
      setIsProcessingJob(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 p-8">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-4xl font-bold text-white mb-8 background-blur">InterviewMe- CV-Score Module</h1>
        
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
                  type="file" 
                  className="hidden" 
                  onChange={handleFileUpload} 
                  accept=".pdf,.doc,.docx"
                  disabled={isUploading || isPreprocessing} 
                />
              </label>
            </div>
            {preprocessedData && (
              <div className="mt-4 text-white">
                <h3 className="font-semibold mb-2">Extracted Information:</h3>
                <div className="bg-gray-700 p-3 rounded text-sm space-y-2">
                  <div><span className="text-gray-400">Name:</span> {preprocessedData.entities.name}</div>
                  <div>
                    <span className="text-gray-400">Email:</span>{' '}
                    {Array.isArray(preprocessedData.entities.email) 
                      ? preprocessedData.entities.email.join(', ')
                      : preprocessedData.entities.email || 'N/A'}
                  </div>
                  <div><span className="text-gray-400">Phone:</span> {preprocessedData.entities.phone}</div>
                  <div><span className="text-gray-400">Skills:</span> {preprocessedData.entities.skills.join(', ')}</div>
                  <div className="text-gray-400">Education:</div>
                  <div className="pl-4">
                    {preprocessedData.entities.education.map((edu, index) => (
                      <div key={index} className="mb-1">
                        {[
                          edu.degree && `${edu.degree}`,
                          edu.specialization && `in ${edu.specialization}`,
                          edu.institution && `from ${edu.institution}`,
                          edu.location && `(${edu.location})`,
                          edu.graduation_date || edu.year
                        ].filter(Boolean).join(' ')}
                      </div>
                    ))}
                  </div>
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
            {jobAnalysis && (
              <div className="mt-4 text-white">
                <h3 className="font-semibold mb-2">Analysis:</h3>
                <pre className="whitespace-pre-wrap text-sm bg-gray-700 p-2 rounded">
                  {JSON.stringify(jobAnalysis, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </div>

        {/* AI Models Section */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {aiModels.map((model) => (
            <div
              key={model.id}
              className="bg-gray-800 p-6 rounded-lg border-2 border-opacity-50 transition-all hover:border-opacity-100"
              style={{ borderColor: model.color }}
            >
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-xl font-semibold text-white flex items-center">
                  <FiCpu className="mr-2" style={{ color: model.color }} />
                  {model.name}
                </h3>
              </div>

              {uploadedCV && (
                <button
                  onClick={() => processWithModel(model.id)}
                  disabled={processingModels[model.id]}
                  className="w-full py-2 px-4 rounded-lg text-white font-medium transition-all"
                  style={{
                    backgroundColor: processingModels[model.id] ? '#374151' : model.color,
                    opacity: processingModels[model.id] ? 0.7 : 1
                  }}
                >
                  {processingModels[model.id] ? 'Processing...' : 'Analyze CV'}
                </button>
              )}

              {modelScores[model.id] && (
                <div className="mt-4 space-y-2">
                  {Object.entries(modelScores[model.id]).map(([key, value]) => (
                    <div key={key} className="flex justify-between items-center">
                      <span className="text-gray-300 capitalize">{key.replace('_', ' ')}</span>
                      <span className="text-white font-semibold">{value}%</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>

        {error && (
          <div className="mt-4 p-4 bg-red-900/50 border border-red-500 rounded-lg text-red-200">
            {error}
          </div>
        )}
      </div>
    </div>
  );
}
