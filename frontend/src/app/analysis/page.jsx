'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';  // Add this import
import dynamic from 'next/dynamic';
import axios from 'axios';
import { FaCheckCircle, FaBriefcase, FaGraduationCap, FaTools, FaChartLine, 
         FaLanguage, FaCertificate, FaProjectDiagram, FaUserTie, FaArrowLeft, FaArrowRight, FaSpinner } from 'react-icons/fa';

const AnalysisPage = () => {
  const [isMounted, setIsMounted] = useState(false);
  const router = useRouter();
  const [error, setError] = useState(null);
  const [showContent, setShowContent] = useState(false);
  const [analysisData, setAnalysisData] = useState(null);
  const [pdfData, setPdfData] = useState(null);
  const [projectsData, setProjectsData] = useState(null);
  const [isLoadingProjects, setIsLoadingProjects] = useState(true);
  const [experienceAnalysis, setExperienceAnalysis] = useState(null);
  const [isLoadingExperience, setIsLoadingExperience] = useState(true);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!isMounted) return;
    
    try {
      const storedData = localStorage.getItem('scoringResults');
      if (!storedData) {
        throw new Error('No scoring data found');
      }

      const parsedData = JSON.parse(storedData);
      console.log('Retrieved stored data:', parsedData);

      if (!parsedData.scores || !parsedData.resumeData) {
        throw new Error('Invalid scoring data');
      }

      // Simplified ID cleaning functions - just get timestamp
      const getTimestamp = (id) => {
        if (!id) return Date.now();
        const match = id.match(/\d+/);
        return match ? match[0] : Date.now();
      };

      // Clean IDs to ensure single prefix
      const cleanResumeId = (id) => `resume_${getTimestamp(id)}`;
      const cleanJobId = (id) => `job_${getTimestamp(id)}`;

      // Set analysis data with simple timestamp IDs
      setAnalysisData({
        overall_score: Math.round(parsedData.scores.overall_score || 0),
        technical_score: Math.round(parsedData.scores.detailed_scores?.technical_skills || 0),
        soft_skills_score: Math.round(parsedData.scores.detailed_scores?.soft_skills || 0),
        experience_score: Math.round(parsedData.scores.experience_match || 0),
        education_score: Math.round(parsedData.scores.education_match || 0),
        education: parsedData.resumeData.entities.education || [],
        experience: parsedData.resumeData.entities.experience || [],
        projects: parsedData.resumeData.entities.projects || [],
        skills: parsedData.resumeData.entities.skills || [],
        recommendations: parsedData.scores.feedback || ['No recommendations available'],
        matchedSkills: parsedData.scores.matched_skills || { exact_matches: [], partial_matches: [] },
        soft_skills_analysis: parsedData.scores.soft_skills_analysis || {
          score: 0,
          extracted_skills: [],
          experience_based_skills: []
        },
        filename: parsedData.resumeData.filename, // Add this line
        resumeId: cleanResumeId(parsedData.resumeId || parsedData.resumeData.id || parsedData.resumeData.filename),
        jobId: cleanJobId(parsedData.jobId || parsedData.jobData.id),
      });

      setShowContent(true);

    } catch (error) {
      console.error('Error:', error);
      setError('Unable to load analysis data');
      setTimeout(() => router.push('/CVScoring'), 3000);
    }
  }, [isMounted, router]);

  useEffect(() => {
    const fetchProjectsData = async () => {
      try {
        if (!analysisData?.resumeId || !analysisData?.jobId) {
          console.log('Waiting for IDs...');
          return;
        }

        // Use IDs exactly as they are - no additional processing
        const requestData = {
          resume_id: analysisData.resumeId,
          job_id: analysisData.jobId
        };

        console.log('Sending request with IDs:', requestData);

        const response = await axios.post(
          `${process.env.NEXT_PUBLIC_API_URL}/analyze-projects`,
          requestData
        );

        if (response.data?.projects) {
          setProjectsData(response.data.projects);
        } else {
          throw new Error('No projects data in response');
        }
      } catch (error) {
        console.error('Project analysis error:', error.response?.data || error.message);
        setProjectsData([]);
      } finally {
        setIsLoadingProjects(false);
      }
    };

    if (isMounted && analysisData) {
      fetchProjectsData();
    }
  }, [isMounted, analysisData]);

  useEffect(() => {
    const fetchExperienceData = async () => {
      try {
        if (!analysisData?.resumeId || !analysisData?.jobId) {
          console.log('Waiting for IDs...');
          return;
        }

        const requestData = {
          resume_id: analysisData.resumeId,
          job_id: analysisData.jobId
        };

        console.log('Fetching experience data with:', requestData);

        const response = await axios.post(
          `${process.env.NEXT_PUBLIC_API_URL}/analyze-experience`,
          requestData
        );

        if (response.data) {
          console.log('Experience analysis received:', response.data);
          setExperienceAnalysis(response.data);
        }
      } catch (error) {
        console.error('Experience analysis error:', error);
        setExperienceAnalysis({
          total_years: 0,
          roles: [],
          skills_gained: [],
          job_fit_analysis: "Failed to analyze experience",
          responsibility_map: {}
        });
      } finally {
        setIsLoadingExperience(false);
      }
    };

    if (isMounted && analysisData) {
      fetchExperienceData();
    }
  }, [isMounted, analysisData]);

  const handlePdfUpload = async (file) => {
    try {
      const formData = new FormData();
      formData.append('file', file);

      const response = await axios.post(`${process.env.NEXT_PUBLIC_API_URL}/upload`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      if (response.data.filename) {
        setPdfData(`${process.env.NEXT_PUBLIC_API_URL}/pdf/display?filename=${response.data.filename}`);
      }
    } catch (error) {
      console.error('Error uploading PDF:', error);
    }
  };

  const handleViewPDF = () => {
    try {
      // Get filename from stored data
      const storedData = localStorage.getItem('scoringResults');
      if (!storedData) {
        console.error('No resume data found');
        return;
      }

      const parsedData = JSON.parse(storedData);
      const filename = parsedData.resumeData?.filename;
      
      if (filename) {
        // Just use the filename as stored, which should already include the "upload-" prefix
        setPdfData(`${process.env.NEXT_PUBLIC_API_URL}/pdf/display?filename=${filename}`);
      } else {
        console.error('No filename found in stored data');
      }
    } catch (error) {
      console.error('Error viewing PDF:', error);
    }
  };

  if (!isMounted) return null;

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 text-white p-8 flex items-center justify-center">
        <div className="bg-red-500/10 border border-red-500 rounded-lg p-4">
          {error}
        </div>
      </div>
    );
  }

  const EducationSection = ({ education }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaGraduationCap className="mr-2 text-green-400" />
        Education
      </h3>
      <div className="space-y-4">
        {education.map((edu, index) => (
          <div key={index} className="p-4 bg-gray-700 rounded">
            <h4 className="font-semibold text-green-400">{edu.degree}</h4>
            <p className="text-gray-300">{edu.institution}</p>
            <p className="text-gray-400">{edu.year}</p>
            {edu.specialization && (
              <p className="text-gray-300 mt-1">Specialization: {edu.specialization}</p>
            )}
          </div>
        ))}
      </div>
    </div>
  );

  const ExperienceSection = ({ experience, totalYears, jobFitAnalysis }) => {
    // Ensure we have valid experience data
    const roles = experience?.map(exp => ({
      title: exp.title || 'Untitled Role',
      company: exp.company || 'Company Not Specified',
      duration: exp.duration || 'Duration Not Specified',
      skills: exp.skills || [],
      responsibilities: exp.responsibilities || []
    })) || [];
  
    const [currentRole, setCurrentRole] = useState(0);
    const [isTransitioning, setIsTransitioning] = useState(false);
  
    // Return early if no experience data
    if (!roles || roles.length === 0) {
      return (
        <div className="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700 shadow-xl">
          <div className="flex items-center justify-center h-40">
            <div className="text-center">
              <FaBriefcase className="text-4xl text-gray-600 mx-auto mb-4" />
              <p className="text-gray-400 text-lg">No experience found in resume</p>
            </div>
          </div>
        </div>
      );
    }
  
    const handleRoleChange = (direction) => {
      if (roles.length <= 1) return; // Don't animate if only one role
      
      setIsTransitioning(true);
      setTimeout(() => {
        if (direction === 'next') {
          setCurrentRole((prev) => (prev + 1) % roles.length);
        } else {
          setCurrentRole((prev) => (prev - 1 + roles.length) % roles.length);
        }
        setIsTransitioning(false);
      }, 300);
    };
  
    const role = roles[currentRole];
    
    // Ensure we have a valid role object
    if (!role) {
      return (
        <div className="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700 shadow-xl">
          <div className="flex items-center justify-center h-40">
            <div className="text-center">
              <FaBriefcase className="text-4xl text-red-400 mx-auto mb-4" />
              <p className="text-red-400">Error loading experience data</p>
            </div>
          </div>
        </div>
      );
    }
  
    return (
      <div className="bg-gradient-to-br from-purple-900/50 to-gray-900 rounded-xl p-8 border border-purple-700/30 shadow-2xl">
        {/* Add total years display */}
        <div className="mb-4 text-center">
          <span className="text-purple-400 text-lg">Total Experience: </span>
          <span className="text-2xl font-bold text-white">{totalYears} years</span>
        </div>

        <div className="flex justify-between items-center mb-6">
          <h3 className="text-2xl font-bold flex items-center text-purple-400">
            <FaBriefcase className="mr-3" />
            Professional Experience
          </h3>
          {roles.length > 1 && (
            <div className="flex items-center gap-3">
              <span className="text-sm text-gray-400">
                {currentRole + 1} of {roles.length}
              </span>
              <div className="flex gap-2">
                <button
                  onClick={() => handleRoleChange('prev')}
                  className="p-2 rounded-lg bg-gray-700/50 hover:bg-purple-600/30 transition-all duration-300
                           border border-gray-600 hover:border-purple-500 group"
                  aria-label="Previous role"
                >
                  <FaArrowLeft className="text-gray-400 group-hover:text-purple-400 transition-colors" />
                </button>
                <button
                  onClick={() => handleRoleChange('next')}
                  className="p-2 rounded-lg bg-gray-700/50 hover:bg-purple-600/30 transition-all duration-300
                           border border-gray-600 hover:border-purple-500 group"
                  aria-label="Next role"
                >
                  <FaArrowRight className="text-gray-400 group-hover:text-purple-400 transition-colors" />
                </button>
              </div>
            </div>
          )}
        </div>
  
        <div className={`transform transition-all duration-300 ${isTransitioning ? 'opacity-0 scale-95' : 'opacity-100 scale-100'}`}>
          <div className="bg-gray-800/40 backdrop-blur-sm rounded-lg p-6 space-y-6">
            <div className="border-b border-purple-600/30 pb-4">
              <h4 className="text-2xl font-semibold text-purple-400 mb-2">
                {role.title}
              </h4>
              <p className="text-gray-300 text-lg">{role.company}</p>
              <p className="text-gray-400">{role.duration}</p>
            </div>
  
            {Array.isArray(role.skills) && role.skills.length > 0 && (
              <div className="space-y-2">
                <h5 className="text-lg font-semibold text-blue-400 flex items-center">
                  <span className="mr-2">💡</span>
                  Skills Applied
                </h5>
                <div className="flex flex-wrap gap-2">
                  {role.skills.map((skill, i) => (
                    <span
                      key={i}
                      className="px-4 py-2 bg-blue-500/10 border border-blue-500/30 rounded-full text-sm
                               text-blue-300 hover:bg-blue-500/20 transition-colors duration-300"
                    >
                      {skill}
                    </span>
                  ))}
                </div>
              </div>
            )}
  
            {Array.isArray(role.responsibilities) && role.responsibilities.length > 0 && (
              <div className="space-y-2">
                <h5 className="text-lg font-semibold text-green-400 flex items-center">
                  <span className="mr-2">✓</span>
                  Key Responsibilities
                </h5>
                <div className="space-y-2">
                  {role.responsibilities.map((resp, i) => (
                    <div
                      key={i}
                      className="p-3 bg-green-500/10 border border-green-500/30 rounded-lg text-sm
                               text-gray-300 hover:bg-green-500/20 transition-colors duration-300"
                    >
                      {resp}
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
  
          {roles.length > 1 && (
            <div className="mt-6 flex justify-center">
              <div className="flex gap-2">
                {roles.map((_, index) => (
                  <button
                    key={index}
                    onClick={() => setCurrentRole(index)}
                    className={`transition-all duration-300 ${
                      index === currentRole
                        ? 'w-8 h-2 bg-purple-400'
                        : 'w-2 h-2 bg-gray-600 hover:bg-gray-500'
                    } rounded-full`}
                    aria-label={`Go to role ${index + 1}`}
                  />
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Add job fit analysis at the bottom */}
        {jobFitAnalysis && (
          <div className="mt-6 p-4 bg-purple-500/10 border border-purple-500/30 rounded-lg">
            <h5 className="text-lg font-semibold text-purple-400 mb-2">Job Fit Analysis</h5>
            <p className="text-gray-300 leading-relaxed">{jobFitAnalysis}</p>
          </div>
        )}
      </div>
    );
  };

  // Fix ProjectsSection component
  const ProjectsSection = ({ projects }) => {
    const [currentProject, setCurrentProject] = useState(0);
    const [isTransitioning, setIsTransitioning] = useState(false);
  
    if (isLoadingProjects) {
      return (
        <div className="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700 shadow-xl">
          <div className="flex items-center justify-center h-40">
            <div className="text-center space-y-4">
              <FaSpinner className="text-4xl text-green-400 mx-auto animate-spin" />
              <p className="text-gray-400 text-lg">Analyzing projects...</p>
            </div>
          </div>
        </div>
      );
    }

    if (!projects || projects.length === 0) {
      return (
        <div className="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700 shadow-xl">
          <div className="flex items-center justify-center h-40">
            <div className="text-center">
              <FaProjectDiagram className="text-4xl text-gray-600 mx-auto mb-4" />
              <p className="text-gray-400 text-lg">No projects found in resume</p>
            </div>
          </div>
        </div>
      );
    }

    const handleProjectChange = (direction) => {
      if (projects.length <= 1) return;
      
      setIsTransitioning(true);
      setTimeout(() => {
        if (direction === 'next') {
          setCurrentProject((prev) => (prev + 1) % projects.length);
        } else {
          setCurrentProject((prev) => (prev - 1 + projects.length) % projects.length);
        }
        setIsTransitioning(false);
      }, 300);
    };

    const project = projects[currentProject];

    return (
      <div className="bg-gradient-to-br from-gray-800 to-gray-900 rounded-xl p-8 border border-gray-700 shadow-2xl">
        <div className="flex justify-between items-center mb-6">
          <h3 className="text-2xl font-bold flex items-center text-green-400">
            <FaProjectDiagram className="mr-3" />
            Project Portfolio
          </h3>
          {projects.length > 1 && (
            <div className="flex items-center gap-3">
              <span className="text-sm text-gray-400">
                {currentProject + 1} of {projects.length}
              </span>
              <div className="flex gap-2">
                <button
                  onClick={() => handleProjectChange('prev')}
                  className="p-2 rounded-lg bg-gray-700/50 hover:bg-gray-600 transition-all duration-300
                           border border-gray-600 hover:border-green-500 group"
                  aria-label="Previous project"
                >
                  <FaArrowLeft className="text-gray-400 group-hover:text-green-400 transition-colors" />
                </button>
                <button
                  onClick={() => handleProjectChange('next')}
                  className="p-2 rounded-lg bg-gray-700/50 hover:bg-gray-600 transition-all duration-300
                           border border-gray-600 hover:border-green-500 group"
                  aria-label="Next project"
                >
                  <FaArrowRight className="text-gray-400 group-hover:text-green-400 transition-colors" />
                </button>
              </div>
            </div>
          )}
        </div>

        <div className={`transform transition-all duration-300 ${isTransitioning ? 'opacity-0 scale-95' : 'opacity-100 scale-100'}`}>
          <div className="bg-gray-700/50 backdrop-blur-sm rounded-lg p-6 space-y-6">
            <div className="border-b border-gray-600 pb-4">
              <h4 className="text-2xl font-semibold text-green-400 mb-3">{project.name}</h4>
              <p className="text-gray-300 leading-relaxed">{project.description}</p>
            </div>

            {project.tech_stack && project.tech_stack.length > 0 && (
              <div className="space-y-2">
                <h5 className="text-lg font-semibold text-blue-400 flex items-center">
                  <span className="mr-2">🛠</span>
                  Tech Stack
                </h5>
                <div className="flex flex-wrap gap-2">
                  {project.tech_stack.map((tech, i) => (
                    <span
                      key={i}
                      className="px-4 py-2 bg-blue-500/10 border border-blue-500/30 rounded-full text-sm
                               text-blue-300 hover:bg-blue-500/20 transition-colors duration-300"
                    >
                      {tech}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  };

  const TechnicalSkillsSection = ({ score, matchedSkills }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaTools className="mr-2 text-green-400" />
        Technical Skills
      </h3>
      <div className="text-4xl font-bold text-green-400 mb-6">
        {score}%
      </div>
      
      {/* Exact Matches */}
      {matchedSkills?.exact_matches?.length > 0 && (
        <div className="mb-6">
          <h4 className="text-lg font-semibold text-green-400 mb-3">Matched Skills With Model</h4>
          <div className="flex flex-wrap gap-2">
            {matchedSkills.exact_matches.map((skill, index) => (
              <span key={index} className="px-3 py-1 bg-green-500/20 border border-green-500 rounded-full text-sm">
                {skill}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  const SoftSkillsSection = ({ score, softSkillsAnalysis }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <h3 className="text-xl font-bold mb-4 flex items-center">
        <FaChartLine className="mr-2 text-green-400" />
        Soft Skills
      </h3>
      <div className="text-4xl font-bold text-green-400 mb-6">
        {score}%
      </div>
      
      {/* Extracted Soft Skills */}
      {softSkillsAnalysis?.extracted_skills?.length > 0 && (
        <div className="mb-6">
          <h4 className="text-lg font-semibold text-blue-400 mb-3">Identified Skills With Model</h4>
          <div className="flex flex-wrap gap-2">
            {softSkillsAnalysis.extracted_skills.map((skill, index) => (
              <span key={index} className="px-3 py-1 bg-blue-500/20 border border-blue-500 rounded-full text-sm">
                {skill}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Experience-based Skills */}
      {softSkillsAnalysis?.experience_based_skills?.length > 0 && (
        <div>
          <h4 className="text-lg font-semibold text-purple-400 mb-3">Experience-Based Skills</h4>
          <div className="flex flex-wrap gap-2">
            {softSkillsAnalysis.experience_based_skills.map((skill, index) => (
              <span key={index} className="px-3 py-1 bg-purple-500/20 border border-purple-500 rounded-full text-sm">
                {skill}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8" suppressHydrationWarning>
      <div className="fixed top-4 right-4 z-50 flex space-x-4">
        {/* Add this button */}
        <Link
          href="/"
          className="px-6 py-3 bg-gray-800/90 hover:bg-purple-600 text-white rounded-lg 
                    font-semibold transition-all duration-300 flex items-center space-x-2
                    border border-purple-500 backdrop-blur-sm"
        >
          <span>↩ Home</span>
        </Link>
        {/* Existing buttons */}
        <button
          onClick={handleViewPDF}
          className="px-6 py-3 bg-gray-800/90 hover:bg-blue-600 text-white rounded-lg 
                    font-semibold transition-all duration-300 flex items-center space-x-2
                    border border-blue-500 backdrop-blur-sm cursor-pointer"
        >
          <span>View Current Resume</span>
        </button>
        <button
          onClick={() => router.push('/CVScoring')}
          className="px-6 py-3 bg-gray-800/90 hover:bg-green-600 text-white rounded-lg 
                    font-semibold transition-all duration-300 flex items-center space-x-2
                    border border-green-500 backdrop-blur-sm"
        >
          <span>↩ Analyze Another Resume</span>
        </button>
      </div>

      {/* Add PDF viewer modal */}
      {pdfData && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-75 p-4">
          <div className="relative w-full max-w-4xl h-[80vh] bg-white rounded-lg">
            <button
              onClick={() => setPdfData(null)}
              className="absolute -top-2 -right-2 w-8 h-8 flex items-center justify-center 
                       bg-red-500 hover:bg-red-600 text-white rounded-full 
                       shadow-lg transform transition-all duration-200 
                       hover:scale-110 focus:outline-none"
              aria-label="Close PDF viewer"
            >
              ✕
            </button>
            <iframe
              src={pdfData}
              className="w-full h-full rounded-lg"
              title="PDF Viewer"
            />
          </div>
        </div>
      )}

      <div className="max-w-6xl mx-auto">
        {showContent && analysisData && (
          <div className="space-y-8">
            {/* Overall Score */}
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
              <h2 className="text-2xl font-bold mb-4 text-green-400">Overall Match Score</h2>
              <div className="text-6xl font-bold text-center text-green-400">
                {analysisData.overall_score}%
              </div>
            </div>

            {/* Detailed Scores */}
            <div className="grid md:grid-cols-2 gap-6">
              <TechnicalSkillsSection 
                score={analysisData.technical_score} 
                matchedSkills={analysisData.matchedSkills}
              />
              <SoftSkillsSection 
                score={analysisData.soft_skills_score}
                softSkillsAnalysis={analysisData.soft_skills_analysis}
              />
            </div>

            <div className="grid md:grid-cols-2 gap-6">
              <EducationSection education={analysisData.education} />
              {isLoadingExperience ? (
                <div className="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700 shadow-xl">
                  <div className="flex items-center justify-center h-40">
                    <div className="text-center space-y-4">
                      <FaSpinner className="text-4xl text-purple-400 mx-auto animate-spin" />
                      <p className="text-gray-400 text-lg">Analyzing experience...</p>
                    </div>
                  </div>
                </div>
              ) : (
                <ExperienceSection 
                  experience={experienceAnalysis?.roles || []}
                  totalYears={experienceAnalysis?.total_years || 0}
                  jobFitAnalysis={experienceAnalysis?.job_fit_analysis}
                />
              )}
            </div>

            <ProjectsSection 
              projects={projectsData} // Remove the fallback to analysisData.projects
            />

            {/* Recommendations - Fixed structure */}
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
              <h3 className="text-xl font-bold mb-4 text-green-400">Recommendations</h3>
              <div className="space-y-3">
                {analysisData.recommendations.map((rec, index) => (
                  <div key={index} className="p-3 bg-gray-700 rounded flex items-center">
                    <FaCheckCircle className="text-green-400 mr-3" />
                    <span>{rec}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default dynamic(() => Promise.resolve(AnalysisPage), {
  ssr: false
});
