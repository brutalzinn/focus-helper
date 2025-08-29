# ♾️ Focus Helper: ATC Edition

A personal focus assistant that simulates an Air Traffic Control (ATC) tower to help you disconnect from the computer and avoid hyperfocus. A tool that every people can be usefull to warning about future crises caused by autism burnout

# About The Project

In the world of deep work, it's easy to lose track of time and fall into a state of hyperfocus, neglecting necessary breaks and other tasks. Focus Helper ATC acts as your personal air traffic controller for your attention.

Instead of managing planes, it manages your "focus flights." You declare a task and a time, and the tower guides you to a safe "landing" when your time is up, reminding you to take a break, switch tasks, or disconnect. This prevents burnout and helps maintain a healthy work-life balance by ensuring you remain in control of your focus, not the other way around.
Getting Started on Ubuntu

This guide will walk you through setting up and running the project on an Ubuntu-based system.
Prerequisites

First, you need to install the necessary build tools and the Go programming language.

    Install Build Tools & Git:
    Open your terminal and run the following command to install git, make, cmake, and the C/C++ compiler toolchain.

    sudo apt update && sudo apt install git make cmake build-essential

    Install Go:
    We recommend installing Go from the official source to avoid environment conflicts.

    # Download the latest version (check go.dev/dl for the newest link)
    wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz

    # Install Go to the standard location
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz

    # Add Go to your system's PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
    source ~/.profile

    To verify the installation, open a new terminal and run go version. You should see the installed version number.

Installation & Setup

With the prerequisites installed, you can now set up the project with a single command.

    Clone the Repository:

    git clone <your-project-repository-url>
    cd <your-project-directory>

    Run the Setup Command:
    This command will automatically download the whisper.cpp dependency (as a Git submodule) and compile the necessary C++ libraries.

    make setup

# Running the Application

After the setup is complete, you can start the Focus Helper ATC tower with a simple command.

    make run

This will compile and run the main Go program. The application is now active and ready to manage your focus flights!
Makefile Commands

The project includes a Makefile to simplify common tasks.

    make setup: Sets up the project for the first time (initializes submodules and builds C++).

    make build: Compiles the production binary.

    make run: Compiles and runs the program.

    make build-debug: Creates a binary for debugging with VS Code.

    make clean: Removes the compiled files.

# Development & Debugging

For development, it is highly recommended to use the native .deb package of Visual Studio Code. The project includes the necessary configuration (.vscode/launch.json and .vscode/tasks.json) to enable one-click debugging (F5), which automatically handles the complex CGo compilation process.


# Contributing & License

This project is currently in development and subject to change. Please file feature requests and bugs in the GitHub issues. The license is Apache 2 so feel free to redistribute. Redistributions in either source code or binary form must reproduce the copyright notice, and please link back to this repository for more information:


    whisper.cpp
    https://github.com/ggerganov/whisper.cpp
    Copyright (c) The ggml authors

    ffmpeg
    https://ffmpeg.org/
    Copyright (c) the FFmpeg developers

This software links to static libraries of whisper.cpp licensed under the MIT License. This software links to static libraries of ffmpeg licensed under the LGPL 2.1 License.