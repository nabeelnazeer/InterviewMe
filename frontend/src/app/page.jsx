'use client';
import { motion } from 'framer-motion';
import Link from 'next/link';
import { FaRocket, FaChartLine, FaUserTie, FaRegLightbulb } from 'react-icons/fa';
import { useEffect, useState } from 'react';

export default function Home() {
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  const fadeInUp = {
    initial: { opacity: 0, y: 60 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.6 }
  };

  const features = [
    {
      icon: <FaRocket className="text-4xl text-blue-400" />,
      title: "Instant Analysis",
      description: "Get immediate insights on your resume's strengths and areas for improvement."
    },
    {
      icon: <FaChartLine className="text-4xl text-green-400" />,
      title: "Detailed Scoring",
      description: "Comprehensive evaluation of technical skills, experience, and qualifications."
    },
    {
      icon: <FaUserTie className="text-4xl text-purple-400" />,
      title: "Career Guidance",
      description: "Receive personalized recommendations to enhance your professional profile."
    },
    {
      icon: <FaRegLightbulb className="text-4xl text-yellow-400" />,
      title: "Smart Matching",
      description: "AI-powered matching with job requirements and industry standards."
    }
  ];

  return (
    <div className="min-h-screen bg-gray-900 text-white" suppressHydrationWarning>
      {isClient ? (
        <>
          {/* Hero Section */}
          <motion.section 
            className="relative h-screen flex items-center justify-center overflow-hidden"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 1 }}
          >
            <div className="absolute inset-0 z-0">
              <div className="absolute inset-0 bg-gradient-to-br from-blue-500/20 to-purple-500/20" />
              <div className="absolute inset-0 backdrop-blur-3xl" />
            </div>
            
            <div className="container mx-auto px-6 z-10 text-center">
              <motion.h1 
                className="text-4xl md:text-6xl font-bold mb-6 bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent"
                {...fadeInUp}
              >
                Elevate Your Career Journey
              </motion.h1>
              
              <motion.p 
                className="text-xl md:text-2xl mb-8 text-gray-300"
                {...fadeInUp}
                transition={{ delay: 0.2 }}
              >
                Transform your resume with AI-powered analysis and recommendations
              </motion.p>
              
              <motion.div
                {...fadeInUp}
                transition={{ delay: 0.4 }}
              >
                <Link 
                  href="/CVScoring" 
                  className="px-8 py-4 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full 
                             text-white font-semibold text-lg hover:scale-105 transform transition-all 
                             duration-300 inline-block hover:shadow-lg hover:shadow-purple-500/25"
                >
                  Analyze Your Resume
                </Link>
              </motion.div>
            </div>
          </motion.section>

          {/* Features Section */}
          <section className="py-20 bg-gray-800/50">
            <div className="container mx-auto px-6">
              <motion.div 
                className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8"
                initial={{ opacity: 0, y: 40 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.8 }}
                viewport={{ once: true }}
              >
                {features.map((feature, index) => (
                  <motion.div
                    key={index}
                    className="p-6 rounded-xl bg-gray-800/50 backdrop-blur-sm border border-gray-700/50 
                               hover:border-gray-600 transition-all duration-300 hover:shadow-xl 
                               hover:shadow-purple-500/10"
                    whileHover={{ y: -5 }}
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.1 }}
                    viewport={{ once: true }}
                  >
                    <div className="mb-4">{feature.icon}</div>
                    <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
                    <p className="text-gray-400">{feature.description}</p>
                  </motion.div>
                ))}
              </motion.div>
            </div>
          </section>

          {/* Call to Action Section */}
          <motion.section 
            className="py-20 text-center"
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            transition={{ duration: 0.8 }}
            viewport={{ once: true }}
          >
            <div className="container mx-auto px-6">
              <h2 className="text-3xl md:text-4xl font-bold mb-6">
                Ready to Take Your Career to the Next Level?
              </h2>
              <p className="text-xl text-gray-400 mb-8">
                Join thousands of professionals who trust our AI-powered resume analysis
              </p>
              <Link 
                href="/CVScoring"
                className="px-8 py-4 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full 
                           text-white font-semibold text-lg hover:scale-105 transform transition-all 
                           duration-300 inline-block hover:shadow-lg hover:shadow-purple-500/25"
              >
                Get Started Now
              </Link>
            </div>
          </motion.section>
        </>
      ) : (
        // Show a simple loading state or static content during SSR
        <div className="h-screen flex items-center justify-center">
          <div className="text-2xl">Loading...</div>
        </div>
      )}
    </div>
  );
}
