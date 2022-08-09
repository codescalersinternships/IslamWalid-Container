package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

const ProcessesLimit = "10"

func main() {
    if len(os.Args) < 2 {
        panic("Few arguments")
    }

    switch os.Args[1] {
    case "run":
        Run(os.Args[2], os.Args[3:])
    case "child":
        Child(os.Args[2], os.Args[3:])
    default:
        panic("Bad usage")
    }
}

// Start the parent process of the container
func Run(command string, args []string) {
    cmd := exec.Command("/proc/self/exe", append([]string{"child", command}, args...)...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Credential: &syscall.Credential{
            Uid: uint32(os.Getuid()),
            Gid: uint32(os.Getgid()),
        },
        Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
        Unshareflags: syscall.CLONE_NEWNS,
        UidMappings:  []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getuid(), Size: 1}},
        GidMappings:  []syscall.SysProcIDMap{{ContainerID: 0, HostID: os.Getgid(), Size: 1}},
    }

    must(cmd.Run())
}

// Start a child process to execute given command
func Child(command string, args []string) {
    fmt.Printf("Running %v with pid %d\n", append([]string{command}, args...), os.Getpid())

    // Setup the new root
    must(syscall.Sethostname([]byte("container")))
    must(syscall.Chroot("./rootfs"))
    must(os.Chdir("/"))
    must(syscall.Mount("proc", "proc", "proc", 0, ""))
    must(syscall.Mount("dev", "dev", "tmpfs", 0, ""))

    cmd := exec.Command(command, args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    must(cmd.Run())

    // Clean up
    must(syscall.Unmount("/proc", 0))
    must(syscall.Unmount("/dev", 0))
}

func CreateCgroup() {
    containerPidsDir := "/sys/fs/cgroup/pids"
    os.Mkdir(containerPidsDir, 0755)

    must(os.WriteFile(path.Join(containerPidsDir, "pids.max"), []byte(ProcessesLimit), 0700))
    must(os.WriteFile(path.Join(containerPidsDir, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

// always panic on errors
func must(err error) {
    if err != nil {
        panic(err)
    }
}
